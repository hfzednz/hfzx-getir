package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcadapter "github.com/nexora/quality-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/quality-service/internal/adapters/http"
	"github.com/nexora/quality-service/internal/adapters/kafka"
	"github.com/nexora/quality-service/internal/adapters/postgres"
	"github.com/nexora/quality-service/internal/app"
	"github.com/nexora/quality-service/internal/app/memory"
	"github.com/nexora/quality-service/internal/config"
	"github.com/nexora/quality-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	deps := &app.Deps{
		Suites: repos.Suites, Runs: repos.Runs, Results: repos.Results, Coverage: repos.Coverage,
		Policies: repos.Policies, Evals: repos.Evals, Certs: repos.Certs, Flaky: repos.Flaky,
		Perf: repos.Perf, Security: repos.Security, Outbox: repos.Outbox,
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log), Runner: repos.Runner, Metrics: repos.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}

	var db *sql.DB
	if cfg.DevMode() {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	} else {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		pg := postgres.NewRepos(db)
		deps.Suites = pg.Suites
		deps.Runs = pg.Runs
		deps.Results = pg.Results
		deps.Coverage = pg.Coverage
		deps.Policies = pg.Policies
		deps.Evals = pg.Evals
		deps.Certs = pg.Certs
		deps.Flaky = pg.Flaky
		deps.Perf = pg.Perf
		deps.Security = pg.Security
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}
	_ = db

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()
	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr: cfg.HTTPAddr, Deps: deps, Limiter: ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute, CORSOrigins: cfg.CORSAllowedOrigins, Log: log,
	})
	go func() {
		log.Info("http.listen", "addr", cfg.HTTPAddr, "devMode", cfg.DevMode())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http.serve", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("shutdown.complete")
}
