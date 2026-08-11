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

	grpcadapter "github.com/nexora/search-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/search-service/internal/adapters/http"
	"github.com/nexora/search-service/internal/adapters/kafka"
	"github.com/nexora/search-service/internal/adapters/postgres"
	"github.com/nexora/search-service/internal/app"
	"github.com/nexora/search-service/internal/app/memory"
	"github.com/nexora/search-service/internal/config"
	"github.com/nexora/search-service/internal/ratelimit"
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
		Docs: repos.Docs, Synonyms: repos.Synonyms, Boosts: repos.Boosts,
		Jobs: repos.Jobs, Trends: repos.Trends, Suggests: repos.Suggests, Outbox: repos.Outbox,
		Lexical: repos.Lexical, Vectors: repos.Vectors,
		Embed: memory.MockEmbed{}, LLM: memory.MockLLM{},
		Catalog: &memory.MockCatalog{}, Inventory: memory.MockInventory{},
		Pricing: memory.MockPricing{}, Reviews: memory.MockReviews{},
		Recs: &memory.MockRecs{}, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
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
		deps.Docs = pg.Docs
		deps.Synonyms = pg.Synonyms
		deps.Boosts = pg.Boosts
		deps.Jobs = pg.Jobs
		deps.Trends = pg.Trends
		deps.Suggests = pg.Suggests
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		}
		return nil
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()
	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr: cfg.HTTPAddr, Deps: deps, Limiter: ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute, CORSOrigins: cfg.CORSAllowedOrigins, Log: log,
		Live:  func(*http.Request) error { return nil },
		Ready: ready,
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
