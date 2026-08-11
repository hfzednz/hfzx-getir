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

	grpcadapter "github.com/nexora/erp-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/erp-service/internal/adapters/http"
	"github.com/nexora/erp-service/internal/adapters/kafka"
	"github.com/nexora/erp-service/internal/adapters/postgres"
	"github.com/nexora/erp-service/internal/app"
	"github.com/nexora/erp-service/internal/app/memory"
	"github.com/nexora/erp-service/internal/config"
	"github.com/nexora/erp-service/internal/ratelimit"
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
		Companies: repos.Companies, Periods: repos.Periods, Accounts: repos.Accounts,
		Journals: repos.Journals, Suppliers: repos.Suppliers, Procurement: repos.Procurement,
		AP: repos.AP, AR: repos.AR, Treasury: repos.Treasury, Budgets: repos.Budgets,
		Assets: repos.Assets, Tax: repos.Tax, Expenses: repos.Expenses,
		Approvals: repos.Approvals, Payroll: repos.Payroll, Outbox: repos.Outbox,
		Ledger: repos.Ledger, Inventory: repos.Inventory, Settlement: repos.Settlement, AI: repos.AI,
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{}, IDs: app.UUIDGen{},
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
		deps.Companies = pg.Companies
		deps.Periods = pg.Periods
		deps.Accounts = pg.Accounts
		deps.Journals = pg.Journals
		deps.Suppliers = pg.Suppliers
		deps.Procurement = pg.Procurement
		deps.AP = pg.AP
		deps.AR = pg.AR
		deps.Treasury = pg.Treasury
		deps.Budgets = pg.Budgets
		deps.Assets = pg.Assets
		deps.Tax = pg.Tax
		deps.Expenses = pg.Expenses
		deps.Approvals = pg.Approvals
		deps.Payroll = pg.Payroll
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
