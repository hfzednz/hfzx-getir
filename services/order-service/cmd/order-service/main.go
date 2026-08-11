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

	httpadapter "github.com/nexora/order-service/internal/adapters/http"
	"github.com/nexora/order-service/internal/adapters/dispatch"
	"github.com/nexora/order-service/internal/adapters/inventory"
	"github.com/nexora/order-service/internal/adapters/kafka"
	"github.com/nexora/order-service/internal/adapters/payment"
	"github.com/nexora/order-service/internal/adapters/postgres"
	"github.com/nexora/order-service/internal/adapters/redis"
	searchadapter "github.com/nexora/order-service/internal/adapters/search"
	"github.com/nexora/order-service/internal/adapters/warehouse"
	"github.com/nexora/order-service/internal/app"
	"github.com/nexora/order-service/internal/app/memory"
	"github.com/nexora/order-service/internal/config"
	"github.com/nexora/order-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	orders, events, sagas, outbox, fulfills, returns, refunds := memory.NewRepos(store)

	searchIdx := searchadapter.NewIndexer(cfg.OpenSearchURL, log)
	rdb := redis.NewClient(cfg.RedisURL, log)
	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)

	deps := &app.Deps{
		Orders:       orders,
		Events:       events,
		Sagas:        sagas,
		Outbox:       outbox,
		Fulfillments: fulfills,
		Returns:      returns,
		Refunds:      refunds,
		Search:       searchIdx, // always use indexer (mem fallback when URL empty)
		Publisher:    publisher,
		Inventory:    memory.NewInventoryClient(),
		Payment:      &memory.PaymentClient{},
		Warehouse:    &memory.WarehouseClient{},
		Dispatch:     &memory.DispatchClient{},
		Clock:        app.SystemClock{},
		IDs:          app.UUIDGen{},
		PlaceLock:    &memory.PlaceLocker{S: store},
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
		deps.Orders = repos.Orders
		deps.Events = repos.Events
		deps.Sagas = repos.Sagas
		deps.Outbox = repos.Outbox
		deps.Fulfillments = repos.Fulfillments
		deps.Returns = repos.Returns
		deps.Refunds = repos.Refunds
		deps.Inventory = inventory.NewClient(cfg.InventoryURL, log)
		deps.Payment = payment.NewClient(cfg.PaymentURL, log)
		deps.Warehouse = warehouse.NewClient(cfg.WarehouseURL, log)
		deps.Dispatch = dispatch.NewClient(cfg.DispatchURL, log)
		deps.PlaceLock = &postgres.AdvisoryPlaceLocker{DB: db}
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
