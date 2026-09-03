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

	httpadapter "github.com/nexora/catalog-service/internal/adapters/http"
	"github.com/nexora/catalog-service/internal/adapters/kafka"
	"github.com/nexora/catalog-service/internal/adapters/media"
	"github.com/nexora/catalog-service/internal/adapters/postgres"
	"github.com/nexora/catalog-service/internal/adapters/redis"
	searchadapter "github.com/nexora/catalog-service/internal/adapters/search"
	"github.com/nexora/catalog-service/internal/app"
	"github.com/nexora/catalog-service/internal/app/memory"
	"github.com/nexora/catalog-service/internal/config"
	"github.com/nexora/catalog-service/internal/ratelimit"
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
	products, variants, skus, categories, brands, attributes, locales, seo, mediaRepo, bundles, relations, versions, workflow, importJobs, compliance := memory.NewRepos(store)

	searchIdx := searchadapter.NewIndexer(cfg.OpenSearchURL, log)

	deps := &app.Deps{
		Products:    products,
		Variants:    variants,
		SKUs:        skus,
		Categories:  categories,
		Brands:      brands,
		Attributes:  attributes,
		Locales:     locales,
		SEO:         seo,
		Media:       mediaRepo,
		Bundles:     bundles,
		Relations:   relations,
		Versions:    versions,
		Workflow:    workflow,
		ImportJobs:  importJobs,
		Compliance:  compliance,
		Suppliers:   &memory.SupplierRepo{S: store},
		Search:      searchIdx,
		Events:      publisher,
		MediaClient: media.NewClient(cfg.MediaServiceURL, log),
		AI:          memory.AIClient{},
		Clock:       app.SystemClock{},
		IDs:         app.UUIDGen{},
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
		deps.Products = repos.Products
		deps.Variants = repos.Variants
		deps.SKUs = repos.SKUs
		deps.Categories = repos.Categories
		deps.Brands = repos.Brands
		deps.Attributes = repos.Attributes
		deps.Locales = repos.Locales
		deps.SEO = repos.SEO
		deps.Media = repos.Media
		deps.Bundles = repos.Bundles
		deps.Relations = repos.Relations
		deps.Versions = repos.Versions
		deps.Workflow = repos.Workflow
		deps.ImportJobs = repos.ImportJobs
		deps.Compliance = repos.Compliance
		deps.Suppliers = repos.Suppliers
		deps.Search = searchIdx
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}
	rdb := redis.NewClient(cfg.RedisURL, log)

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
	_ = publisher.Close()
	if err := rdb.Close(); err != nil {
		log.Error("redis.close", "err", err)
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
