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

	httpadapter "github.com/nexora/inventory-service/internal/adapters/http"
	"github.com/nexora/inventory-service/internal/adapters/kafka"
	"github.com/nexora/inventory-service/internal/adapters/postgres"
	"github.com/nexora/inventory-service/internal/adapters/redis"
	searchadapter "github.com/nexora/inventory-service/internal/adapters/search"
	"github.com/nexora/inventory-service/internal/app"
	"github.com/nexora/inventory-service/internal/app/memory"
	"github.com/nexora/inventory-service/internal/config"
	"github.com/nexora/inventory-service/internal/ratelimit"
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
	wh, loc, bal, lots, res, mov, xfer, counts, rets, fc := memory.NewRepos(store)

	searchIdx := searchadapter.NewIndexer(cfg.OpenSearchURL, log)

	holds, err := redis.NewClient(cfg.RedisURL, log)
	if err != nil {
		log.Error("redis.open", "err", err)
		os.Exit(1)
	}
	holds.TTL = cfg.SoftReserveTTL

	deps := &app.Deps{
		Warehouses:     wh,
		Locations:      loc,
		Balances:       bal,
		Lots:           lots,
		Reservations:   res,
		Movements:      mov,
		Transfers:      xfer,
		Counts:         counts,
		Returns:        rets,
		Forecasts:      fc,
		Search:         &memory.SearchIndexer{S: store},
		Events:         publisher,
		AI:             memory.ForecastAIClient{},
		Idempotency:    &memory.IdempotencyStore{S: store},
		Locker:         &memory.StockLocker{S: store},
		Clock:          app.SystemClock{},
		IDs:            app.UUIDGen{},
		SoftReserveTTL: cfg.SoftReserveTTL,
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
		repos := postgres.NewRepos(db)
		deps.Warehouses = repos.Warehouses
		deps.Locations = repos.Locations
		deps.Balances = repos.Balances
		deps.Lots = repos.Lots
		deps.Reservations = repos.Reservations
		deps.Movements = repos.Movements
		deps.Transfers = repos.Transfers
		deps.Counts = repos.Counts
		deps.Returns = repos.Returns
		deps.Forecasts = repos.Forecasts
		deps.Search = searchIdx
		deps.Idempotency = &postgres.IdempotencyStore{DB: db, TTL: 24 * time.Hour}
		deps.Locker = &postgres.AdvisoryStockLocker{DB: db}
		log.Info("boot.database", "driver", "pgx", "repos", "postgres", "redis", cfg.RedisURL != "")
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
			return holds.Ping(ctx)
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
	_ = publisher.Close()
	_ = holds.Close()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
