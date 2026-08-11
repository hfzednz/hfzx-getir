package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcadapter "github.com/nexora/autonomy-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/autonomy-service/internal/adapters/http"
	"github.com/nexora/autonomy-service/internal/adapters/kafka"
	"github.com/nexora/autonomy-service/internal/adapters/postgres"
	"github.com/nexora/autonomy-service/internal/app"
	"github.com/nexora/autonomy-service/internal/app/memory"
	"github.com/nexora/autonomy-service/internal/config"
	"github.com/nexora/autonomy-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}
	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)
	if cfg.DevMode() {
		publisher.AllowNoopWithoutBrokers = true
	} else if len(cfg.KafkaBrokers) == 0 {
		log.Error("boot.kafka", "err", "KAFKA_BROKERS required when DATABASE_URL is set")
		os.Exit(1)
	}
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	deps := &app.Deps{
		Audits: repos.Audits, Weaknesses: repos.Weaknesses, Heals: repos.Heals,
		Reviews: repos.Reviews, Evolution: repos.Evolution, Releases: repos.Releases,
		Governance: repos.Governance, Assistants: repos.Assistants, Teams: repos.Teams,
		Dependencies: repos.Dependencies, Genesis: repos.Genesis, Outbox: repos.Outbox,
		Publisher: publisher,
		Hyperscale: repos.Hyperscale, PlatformOps: repos.Platform, Quality: repos.Quality,
		Security: repos.Security, LiveOps: repos.LiveOps, Metrics: repos.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
	if !cfg.DevMode() {
		db, err := postgres.Open(cfg.DatabaseURL)
		if err != nil {
			log.Error("postgres.open", "err", err)
			os.Exit(1)
		}
		pg := postgres.NewRepos(db)
		deps.Audits = pg.Audits
		deps.Weaknesses = pg.Weaknesses
		deps.Heals = pg.Heals
		deps.Reviews = pg.Reviews
		deps.Evolution = pg.Evolution
		deps.Releases = pg.Releases
		deps.Governance = pg.Governance
		deps.Assistants = pg.Assistants
		deps.Teams = pg.Teams
		deps.Dependencies = pg.Dependencies
		deps.Genesis = pg.Genesis
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "adapter", "postgres", "kafkaBrokers", len(cfg.KafkaBrokers))
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}
	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()
	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr: cfg.HTTPAddr, Deps: deps, Limiter: ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute, CORSOrigins: cfg.CORSAllowedOrigins, Log: log,
	})
	go func() {
		log.Info("http.listen", "addr", cfg.HTTPAddr)
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
