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

	grpcadapter "github.com/nexora/open-platform-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/open-platform-service/internal/adapters/http"
	"github.com/nexora/open-platform-service/internal/adapters/kafka"
	"github.com/nexora/open-platform-service/internal/adapters/postgres"
	"github.com/nexora/open-platform-service/internal/app"
	"github.com/nexora/open-platform-service/internal/app/memory"
	"github.com/nexora/open-platform-service/internal/config"
	"github.com/nexora/open-platform-service/internal/ratelimit"
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
		Apps: repos.Apps, Keys: repos.Keys, Catalog: repos.Catalog, Versions: repos.Versions,
		Policies: repos.Policies, Webhooks: repos.Webhooks, Deliveries: repos.Deliveries,
		SDKs: repos.SDKs, Integrations: repos.Integrations, Usage: repos.Usage,
		Outbox: repos.Outbox, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		HTTP: repos.HTTP, Identity: repos.Identity, Metrics: repos.Metrics,
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
		deps.Apps = pg.Apps
		deps.Keys = pg.Keys
		deps.Catalog = pg.Catalog
		deps.Versions = pg.Versions
		deps.Policies = pg.Policies
		deps.Webhooks = pg.Webhooks
		deps.Deliveries = pg.Deliveries
		deps.SDKs = pg.SDKs
		deps.Integrations = pg.Integrations
		deps.Usage = pg.Usage
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
