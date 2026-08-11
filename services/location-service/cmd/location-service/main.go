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

	geofenceadapter "github.com/nexora/location-service/internal/adapters/geofence"
	grpcadapter "github.com/nexora/location-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/location-service/internal/adapters/http"
	"github.com/nexora/location-service/internal/adapters/kafka"
	"github.com/nexora/location-service/internal/adapters/maps"
	"github.com/nexora/location-service/internal/adapters/postgres"
	"github.com/nexora/location-service/internal/adapters/redis"
	routingadapter "github.com/nexora/location-service/internal/adapters/routing"
	"github.com/nexora/location-service/internal/adapters/search"
	"github.com/nexora/location-service/internal/app"
	"github.com/nexora/location-service/internal/app/memory"
	"github.com/nexora/location-service/internal/config"
	"github.com/nexora/location-service/internal/ratelimit"
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
		Addresses: repos.Addresses,
		POIs:      repos.POIs,
		History:   repos.History,
		Cache:     repos.Cache,
		Heat:      repos.Heat,
		Outbox:    repos.Outbox,
		Maps:      maps.NewClient(log),
		Geofence:  geofenceadapter.NewClient(cfg.GeofenceURL, log),
		Routing:   routingadapter.NewClient(cfg.RoutingURL, log),
		Search:    search.NewClient(cfg.OpenSearchURL, log),
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
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
		pg := postgres.NewRepos(db)
		deps.Addresses = pg.Addresses
		deps.POIs = pg.POIs
		deps.History = pg.History
		deps.Cache = pg.Cache
		deps.Heat = pg.Heat
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

	var redisPOI *redis.Client
	if cfg.RedisURL != "" {
		rc, err := redis.Open(cfg.RedisURL, log)
		if err != nil {
			log.Error("redis.open", "err", err)
			os.Exit(1)
		}
		redisPOI = rc
		deps.POIs = rc
		log.Info("boot.redis", "adapter", "geo-poi", "nearby", "GEORADIUS/GEOSEARCH")
	} else {
		log.Info("boot.redis", "adapter", "memory-haversine", "reason", "REDIS_URL empty")
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				return err
			}
		}
		if redisPOI != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return redisPOI.Ping(ctx)
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
	if redisPOI != nil {
		_ = redisPOI.Close()
	}
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
