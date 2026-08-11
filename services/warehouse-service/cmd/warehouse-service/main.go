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

	httpadapter "github.com/nexora/warehouse-service/internal/adapters/http"
	invadapter "github.com/nexora/warehouse-service/internal/adapters/inventory"
	"github.com/nexora/warehouse-service/internal/adapters/kafka"
	"github.com/nexora/warehouse-service/internal/adapters/postgres"
	"github.com/nexora/warehouse-service/internal/adapters/redis"
	"github.com/nexora/warehouse-service/internal/app"
	"github.com/nexora/warehouse-service/internal/app/memory"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/config"
	"github.com/nexora/warehouse-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	ff, tasks, picks, packs, disp, stations, wf, eq, qc, labels := memory.NewRepos(store)

	rdb := redis.NewClient(cfg.RedisURL, log)

	var invClient ports.InventoryClient = &memory.InventoryClient{S: store}
	if cfg.InventoryServiceURL != "" {
		invClient = invadapter.NewClient(cfg.InventoryServiceURL, log)
	}

	deps := &app.Deps{
		Fulfillments:     ff,
		Tasks:            tasks,
		Picks:            picks,
		Packs:            packs,
		Dispatches:       disp,
		Stations:         stations,
		Workforce:        wf,
		Equipment:        eq,
		QC:               qc,
		Labels:           labels,
		Inventory:        invClient,
		RouteAI:          memory.RouteOptimizer{},
		Events:           kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:            app.SystemClock{},
		IDs:              app.UUIDGen{},
		WeightToleranceG: cfg.WeightToleranceG,
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
		deps.Fulfillments = pg.Fulfillments
		deps.Tasks = pg.Tasks
		deps.Picks = pg.Picks
		deps.Packs = pg.Packs
		deps.Dispatches = pg.Dispatches
		deps.Stations = pg.Stations
		deps.Workforce = pg.Workforce
		deps.Equipment = pg.Equipment
		deps.QC = pg.QC
		deps.Labels = pg.Labels
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
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
	if db != nil {
		_ = db.Close()
	}
	if err := rdb.Close(); err != nil {
		log.Error("redis.close", "err", err)
	}
	log.Info("shutdown.complete")
}
