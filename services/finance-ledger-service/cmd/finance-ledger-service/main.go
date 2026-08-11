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

	httpadapter "github.com/nexora/finance-ledger-service/internal/adapters/http"
	"github.com/nexora/finance-ledger-service/internal/adapters/kafka"
	"github.com/nexora/finance-ledger-service/internal/adapters/postgres"
	"github.com/nexora/finance-ledger-service/internal/adapters/redis"
	"github.com/nexora/finance-ledger-service/internal/app"
	"github.com/nexora/finance-ledger-service/internal/app/memory"
	"github.com/nexora/finance-ledger-service/internal/config"
	"github.com/nexora/finance-ledger-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	accounts, journals, invoices, taxRules, events, outbox := memory.NewRepos(store)
	rdb := redis.NewClient(cfg.RedisURL, log)

	deps := &app.Deps{
		Accounts:  accounts,
		Journals:  journals,
		Invoices:  invoices,
		TaxRules:  taxRules,
		Events:    events,
		Outbox:    outbox,
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
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
		deps.Accounts = repos.Accounts
		deps.Journals = repos.Journals
		deps.Invoices = repos.Invoices
		deps.TaxRules = repos.TaxRules
		deps.Events = repos.Events
		deps.Outbox = repos.Outbox
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
	_ = rdb.Close()
	log.Info("shutdown.complete")
}
