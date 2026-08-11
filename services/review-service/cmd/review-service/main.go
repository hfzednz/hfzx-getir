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

	grpcadapter "github.com/nexora/review-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/review-service/internal/adapters/http"
	"github.com/nexora/review-service/internal/adapters/kafka"
	"github.com/nexora/review-service/internal/adapters/opensearch"
	"github.com/nexora/review-service/internal/adapters/postgres"
	"github.com/nexora/review-service/internal/app"
	"github.com/nexora/review-service/internal/app/memory"
	"github.com/nexora/review-service/internal/config"
	"github.com/nexora/review-service/internal/ratelimit"
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
		Reviews: repos.Reviews, Ratings: repos.Ratings, Media: repos.Media,
		Interactions: repos.Interactions, Quality: repos.Quality, Moderation: repos.Moderation,
		Trust: repos.Trust, Reputation: repos.Reputation, Outbox: repos.Outbox,
		Orders: memory.MockOrders{}, MediaClient: memory.MockMedia{}, AI: memory.MockAI{},
		Search: repos.Search, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
	if cfg.OpenSearchURL != "" {
		deps.Search = opensearch.NewClient(cfg.OpenSearchURL, log)
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
		deps.Reviews = pg.Reviews
		deps.Ratings = pg.Ratings
		deps.Media = pg.Media
		deps.Interactions = pg.Interactions
		deps.Quality = pg.Quality
		deps.Moderation = pg.Moderation
		deps.Trust = pg.Trust
		deps.Reputation = pg.Reputation
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		}
		return nil
	}

	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr:               cfg.HTTPAddr,
		Deps:               deps,
		Limiter:            ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute,
		CORSOrigins:        cfg.CORSAllowedOrigins,
		Log:                log,
		Live:               func(*http.Request) error { return nil },
		Ready:              ready,
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
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
