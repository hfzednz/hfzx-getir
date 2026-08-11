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

	grpcadapter "github.com/nexora/innovation-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/innovation-service/internal/adapters/http"
	"github.com/nexora/innovation-service/internal/adapters/kafka"
	"github.com/nexora/innovation-service/internal/adapters/postgres"
	"github.com/nexora/innovation-service/internal/app"
	"github.com/nexora/innovation-service/internal/app/memory"
	"github.com/nexora/innovation-service/internal/config"
	"github.com/nexora/innovation-service/internal/ratelimit"
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
		Modules: repos.Modules, Experiments: repos.Experiments, Simulations: repos.Simulations,
		Twins: repos.Twins, Edge: repos.Edge, IoT: repos.IoT, Robots: repos.Robots,
		Assignments: repos.Assignments, Drones: repos.Drones, Blockchain: repos.Blockchain,
		XR: repos.XR, Multimodal: repos.Multimodal, Green: repos.Green, Quantum: repos.Quantum,
		Outbox: repos.Outbox, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		LiveOps: repos.LiveOps, AI: repos.AI, Metrics: repos.Metrics,
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
		deps.Modules = pg.Modules
		deps.Experiments = pg.Experiments
		deps.Simulations = pg.Simulations
		deps.Twins = pg.Twins
		deps.Edge = pg.Edge
		deps.IoT = pg.IoT
		deps.Robots = pg.Robots
		deps.Assignments = pg.Assignments
		deps.Drones = pg.Drones
		deps.Blockchain = pg.Blockchain
		deps.XR = pg.XR
		deps.Multimodal = pg.Multimodal
		deps.Green = pg.Green
		deps.Quantum = pg.Quantum
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
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
