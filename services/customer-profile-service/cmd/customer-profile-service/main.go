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

	httpadapter "github.com/nexora/customer-profile-service/internal/adapters/http"
	"github.com/nexora/customer-profile-service/internal/adapters/kafka"
	"github.com/nexora/customer-profile-service/internal/adapters/postgres"
	redisadapter "github.com/nexora/customer-profile-service/internal/adapters/redis"
	"github.com/nexora/customer-profile-service/internal/adapters/search"
	"github.com/nexora/customer-profile-service/internal/app"
	"github.com/nexora/customer-profile-service/internal/app/memory"
	"github.com/nexora/customer-profile-service/internal/config"
	"github.com/nexora/customer-profile-service/internal/observability"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)
	indexer := search.NewIndexer(cfg.SearchURL, log)
	cache := redisadapter.NewCache(cfg.RedisURL)
	if cache.Connected() {
		log.Info("boot.redis", "connected", true)
	} else if cfg.RedisURL != "" {
		log.Warn("boot.redis", "connected", false, "note", "falling back to in-process cache")
	}

	deps := &app.Deps{
		Profiles:        &memory.ProfileRepo{S: store},
		Addresses:       &memory.AddressRepo{S: store},
		Preferences:     &memory.PreferencesRepo{S: store},
		Tags:            &memory.TagRepo{S: store},
		Households:      &memory.HouseholdRepo{S: store},
		Consents:        &memory.ConsentRepo{S: store},
		CRM:             &memory.CRMRepo{S: store},
		Segments:        &memory.SegmentRepo{S: store},
		Personalization: &memory.PersonalizationRepo{S: store},
		AIModels:        &memory.AIModelRepo{S: store},
		Privacy:         &memory.PrivacyRepo{S: store},
		Activity:        &memory.ActivityRepo{S: store},
		Events:          publisher,
		Media:           memory.NewMediaStore(),
		Zones:           &memory.ZoneValidator{OK: true, ZoneID: "zone-local"},
		Search:          indexer,
		Clock:           app.SystemClock{},
		IDs:             app.UUIDGen{},
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
		deps.Profiles = pg.Profiles
		deps.Addresses = pg.Addresses
		deps.Preferences = pg.Preferences
		deps.Tags = pg.Tags
		deps.Households = pg.Households
		deps.Consents = pg.Consents
		deps.CRM = pg.CRM
		deps.Segments = pg.Segments
		deps.Personalization = pg.Personalization
		deps.AIModels = pg.AIModels
		deps.Privacy = pg.Privacy
		deps.Activity = pg.Activity
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				return err
			}
		}
		if cache.Connected() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return cache.Ping(ctx)
		}
		return nil
	}

	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr:               cfg.HTTPAddr,
		Deps:               deps,
		RateLimitPerMinute: cfg.RateLimitPerMinute,
		CORSOrigins:        cfg.CORSAllowedOrigins,
		Log:                log,
		Live:               func(*http.Request) error { return nil },
		Ready:              ready,
	})

	go func() {
		log.Info("http.listen",
			"addr", cfg.HTTPAddr,
			"grpcAddr", cfg.GRPCAddr,
			"devMode", cfg.DevMode(),
			"metrics", observability.Snapshot(),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http.serve", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	if db != nil {
		_ = db.Close()
	}
	_ = cache.Close()
	log.Info("shutdown.complete")
}
