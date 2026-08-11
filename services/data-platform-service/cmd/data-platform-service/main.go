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

	"github.com/nexora/data-platform-service/internal/adapters/clickhouse"
	grpcadapter "github.com/nexora/data-platform-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/data-platform-service/internal/adapters/http"
	"github.com/nexora/data-platform-service/internal/adapters/kafka"
	"github.com/nexora/data-platform-service/internal/adapters/postgres"
	"github.com/nexora/data-platform-service/internal/app"
	"github.com/nexora/data-platform-service/internal/app/memory"
	"github.com/nexora/data-platform-service/internal/app/ports"
	"github.com/nexora/data-platform-service/internal/config"
	"github.com/nexora/data-platform-service/internal/ratelimit"
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
	var olap ports.OLAPWriter = repos.OLAP
	if cfg.ClickHouseURL != "" {
		olap = clickhouse.NewWriter(cfg.ClickHouseURL, repos.OLAP)
	}

	deps := &app.Deps{
		Schemas: repos.Schemas, Events: repos.Events, Streams: repos.Streams,
		Lake: repos.Lake, Warehouse: repos.Warehouse, Realtime: repos.Realtime,
		Experiments: repos.Experiments, Reports: repos.Reports, Obs: repos.Obs,
		Alerts: repos.Alerts, Catalog: repos.Catalog, Quality: repos.Quality,
		Outbox: repos.Outbox, OLAP: olap, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
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
		deps.Schemas = pg.Schemas
		deps.Events = pg.Events
		deps.Streams = pg.Streams
		deps.Lake = pg.Lake
		deps.Warehouse = pg.Warehouse
		deps.Realtime = pg.Realtime
		deps.Experiments = pg.Experiments
		deps.Reports = pg.Reports
		deps.Obs = pg.Obs
		deps.Alerts = pg.Alerts
		deps.Catalog = pg.Catalog
		deps.Quality = pg.Quality
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
