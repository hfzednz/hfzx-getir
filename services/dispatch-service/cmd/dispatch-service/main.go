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

	geofenceadapter "github.com/nexora/dispatch-service/internal/adapters/geofence"
	grpcadapter "github.com/nexora/dispatch-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/dispatch-service/internal/adapters/http"
	"github.com/nexora/dispatch-service/internal/adapters/kafka"
	"github.com/nexora/dispatch-service/internal/adapters/postgres"
	"github.com/nexora/dispatch-service/internal/adapters/redis"
	routingadapter "github.com/nexora/dispatch-service/internal/adapters/routing"
	trackingadapter "github.com/nexora/dispatch-service/internal/adapters/tracking"
	"github.com/nexora/dispatch-service/internal/app"
	"github.com/nexora/dispatch-service/internal/app/memory"
	"github.com/nexora/dispatch-service/internal/config"
	"github.com/nexora/dispatch-service/internal/ratelimit"
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

	routingURL, trackingURL, geofenceURL := "", "", ""
	if !cfg.DevMode() {
		routingURL = cfg.RoutingBaseURL
		trackingURL = cfg.TrackingBaseURL
		geofenceURL = cfg.GeofenceBaseURL
	}

	deps := &app.Deps{
		Dispatches: repos.Dispatches,
		Couriers:   repos.Couriers,
		Vehicles:   repos.Vehicles,
		Outbox:     repos.Outbox,
		Routing:    routingadapter.NewClient(routingURL, log),
		Tracking:   trackingadapter.NewClient(trackingURL, log),
		Geofence:   geofenceadapter.NewClient(geofenceURL, log),
		Publisher:  kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:      app.SystemClock{},
		IDs:        app.UUIDGen{},
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
		deps.Dispatches = pg.Dispatches
		deps.Couriers = pg.Couriers
		deps.Vehicles = pg.Vehicles
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}
	if cfg.RedisURL != "" {
		if _, err := redis.Open(cfg.RedisURL); err != nil {
			log.Warn("redis.open", "err", err)
		}
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
