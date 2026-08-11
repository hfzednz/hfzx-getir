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

	grpcadapter "github.com/nexora/supplier-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/supplier-service/internal/adapters/http"
	"github.com/nexora/supplier-service/internal/adapters/kafka"
	"github.com/nexora/supplier-service/internal/adapters/postgres"
	"github.com/nexora/supplier-service/internal/app"
	"github.com/nexora/supplier-service/internal/app/memory"
	"github.com/nexora/supplier-service/internal/config"
	"github.com/nexora/supplier-service/internal/ratelimit"
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
		Suppliers: repos.Suppliers, Documents: repos.Documents, Certs: repos.Certs,
		Contracts: repos.Contracts, RFQs: repos.RFQs, Quotes: repos.Quotes, POs: repos.POs,
		Shipments: repos.Shipments, Invoices: repos.Invoices, Sellers: repos.Sellers,
		Listings: repos.Listings, Submissions: repos.Submissions, EDI: repos.EDI,
		Scorecards: repos.Scorecards, Messages: repos.Messages, Changes: repos.Changes,
		Outbox: repos.Outbox, Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		ERP: repos.ERP, Catalog: repos.Catalog, Inventory: repos.Inventory,
		Settlement: repos.Settlement, AI: repos.AI, Metrics: repos.Metrics,
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
		deps.Suppliers = pg.Suppliers
		deps.Documents = pg.Documents
		deps.Certs = pg.Certs
		deps.Contracts = pg.Contracts
		deps.RFQs = pg.RFQs
		deps.Quotes = pg.Quotes
		deps.POs = pg.POs
		deps.Shipments = pg.Shipments
		deps.Invoices = pg.Invoices
		deps.Sellers = pg.Sellers
		deps.Listings = pg.Listings
		deps.Submissions = pg.Submissions
		deps.EDI = pg.EDI
		deps.Scorecards = pg.Scorecards
		deps.Messages = pg.Messages
		deps.Changes = pg.Changes
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
