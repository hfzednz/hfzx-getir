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

	grpcadapter "github.com/nexora/recommendation-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/recommendation-service/internal/adapters/http"
	"github.com/nexora/recommendation-service/internal/adapters/kafka"
	"github.com/nexora/recommendation-service/internal/adapters/postgres"
	"github.com/nexora/recommendation-service/internal/app"
	"github.com/nexora/recommendation-service/internal/app/memory"
	"github.com/nexora/recommendation-service/internal/config"
	"github.com/nexora/recommendation-service/internal/ratelimit"
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
		Features: repos.Features, Signals: repos.Signals, CoOccur: repos.CoOccur, Outbox: repos.Outbox,
		Catalog: memory.MockCatalog{}, Trends: &memory.MockTrends{},
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{}, IDs: app.UUIDGen{},
	}

	var db *sql.DB
	if !cfg.DevMode() {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		pg := postgres.NewRepos(db)
		deps.Features = pg.Features
		deps.Signals = pg.Signals
		deps.CoOccur = pg.CoOccur
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

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
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
