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

	grpcadapter "github.com/nexora/superapp-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/superapp-service/internal/adapters/http"
	"github.com/nexora/superapp-service/internal/adapters/kafka"
	"github.com/nexora/superapp-service/internal/adapters/postgres"
	"github.com/nexora/superapp-service/internal/app"
	"github.com/nexora/superapp-service/internal/app/memory"
	"github.com/nexora/superapp-service/internal/config"
	"github.com/nexora/superapp-service/internal/ratelimit"
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
		Modules: repos.Modules, Manifests: repos.Manifests, Installs: repos.Installs,
		Permissions: repos.Permissions, Listings: repos.Listings, Ratings: repos.Ratings,
		Widgets: repos.Widgets, Monetization: repos.Monetization, Launches: repos.Launches,
		Outbox: repos.Outbox, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		LiveOps: repos.LiveOps, AI: repos.AI, Metrics: repos.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{}, ShellVersion: cfg.ShellVersion,
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
		deps.Manifests = pg.Manifests
		deps.Installs = pg.Installs
		deps.Permissions = pg.Permissions
		deps.Listings = pg.Listings
		deps.Ratings = pg.Ratings
		deps.Widgets = pg.Widgets
		deps.Monetization = pg.Monetization
		deps.Launches = pg.Launches
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
