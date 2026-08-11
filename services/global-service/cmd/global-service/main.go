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

	grpcadapter "github.com/nexora/global-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/global-service/internal/adapters/http"
	"github.com/nexora/global-service/internal/adapters/kafka"
	"github.com/nexora/global-service/internal/adapters/postgres"
	"github.com/nexora/global-service/internal/app"
	"github.com/nexora/global-service/internal/app/memory"
	"github.com/nexora/global-service/internal/config"
	"github.com/nexora/global-service/internal/ratelimit"
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
		Countries: repos.Countries, Places: repos.Places, Languages: repos.Languages,
		Locales: repos.Locales, Translations: repos.Translations, Currencies: repos.Currencies,
		Rates: repos.Rates, Holidays: repos.Holidays, Rules: repos.Rules, Taxes: repos.Taxes,
		Privacy: repos.Privacy, PayAvail: repos.PayAvail, Logistics: repos.Logistics, Legal: repos.Legal,
		Outbox: repos.Outbox, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		FX: repos.FX, AI: repos.AI, Metrics: repos.Metrics, Cache: repos.Cache,
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
		deps.Countries = pg.Countries
		deps.Places = pg.Places
		deps.Languages = pg.Languages
		deps.Locales = pg.Locales
		deps.Translations = pg.Translations
		deps.Currencies = pg.Currencies
		deps.Rates = pg.Rates
		deps.Holidays = pg.Holidays
		deps.Rules = pg.Rules
		deps.Taxes = pg.Taxes
		deps.Privacy = pg.Privacy
		deps.PayAvail = pg.PayAvail
		deps.Logistics = pg.Logistics
		deps.Legal = pg.Legal
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}

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
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
