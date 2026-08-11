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

	grpcadapter "github.com/nexora/cart-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/cart-service/internal/adapters/http"
	"github.com/nexora/cart-service/internal/adapters/inventory"
	"github.com/nexora/cart-service/internal/adapters/kafka"
	"github.com/nexora/cart-service/internal/adapters/postgres"
	"github.com/nexora/cart-service/internal/adapters/pricing"
	"github.com/nexora/cart-service/internal/adapters/redis"
	"github.com/nexora/cart-service/internal/app"
	"github.com/nexora/cart-service/internal/app/memory"
	"github.com/nexora/cart-service/internal/config"
	"github.com/nexora/cart-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	carts, events, outbox, saved := memory.NewRepos(store)
	rdb := redis.NewClient(cfg.RedisURL, log)
	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)

	deps := &app.Deps{
		Carts:     carts,
		Events:    events,
		Outbox:    outbox,
		Saved:     saved,
		Publisher: publisher,
		Pricing:   pricing.NewClient("", log),
		Inventory: inventory.NewClient("", log),
		Recommend: &memory.RecommendClient{},
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
	}

	var db *sql.DB
	if !cfg.DevMode() {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		repos := postgres.NewRepos(db)
		deps.Carts = repos.Carts
		deps.Events = repos.Events
		deps.Outbox = repos.Outbox
		deps.Saved = repos.Saved
		deps.Pricing = pricing.NewClient(cfg.PricingURL, log)
		deps.Inventory = inventory.NewClient(cfg.InventoryURL, log)
		log.Info("boot.database", "adapter", "postgres", "kafkaBrokers", len(cfg.KafkaBrokers))
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				return err
			}
		}
		if cfg.RedisURL != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return rdb.Ping(ctx)
		}
		return nil
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()

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
	if err := publisher.Close(); err != nil {
		log.Error("kafka.close", "err", err)
	}
	if err := rdb.Close(); err != nil {
		log.Error("redis.close", "err", err)
	}
	if db != nil {
		if err := db.Close(); err != nil {
			log.Error("postgres.close", "err", err)
		}
	}
	log.Info("shutdown.complete")
}
