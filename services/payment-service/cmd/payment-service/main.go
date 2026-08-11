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

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/adapters/fraud"
	httpadapter "github.com/nexora/payment-service/internal/adapters/http"
	"github.com/nexora/payment-service/internal/adapters/kafka"
	"github.com/nexora/payment-service/internal/adapters/ledger"
	"github.com/nexora/payment-service/internal/adapters/postgres"
	"github.com/nexora/payment-service/internal/adapters/psp"
	"github.com/nexora/payment-service/internal/adapters/wallet"
	"github.com/nexora/payment-service/internal/app"
	"github.com/nexora/payment-service/internal/app/memory"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/config"
	"github.com/nexora/payment-service/internal/domain"
	"github.com/nexora/payment-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	memIntents, memOutbox := memory.NewRepos(store)
	var intents ports.IntentRepo = memIntents
	var outbox ports.OutboxRepository = memOutbox

	var db *sql.DB
	if !cfg.DevMode() {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		repos := postgres.NewRepos(db)
		intents = repos.Intents
		outbox = repos.Outbox
		log.Info("boot.database", "adapter", "postgres")
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

	var router ports.PSPClient
	if stripe, err := psp.NewStripeFromEnv(); err == nil {
		// Production path when Stripe is configured: never attach MockPSP as failover.
		router = psp.NewFailover(stripe)
		log.Info("psp.primary", "provider", "stripe")
	} else if !cfg.DevMode() {
		log.Error("psp.required", "err", err)
		os.Exit(1)
	} else {
		primary := psp.NewMock("mock_primary")
		secondary := psp.NewMock("mock_failover")
		router = psp.NewFailover(primary, secondary)
		log.Info("psp.dev_mock", "reason", err.Error())
	}

	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)
	fraudClient := fraud.NewClient(cfg.AIPlatformURL, log)
	walletClient := wallet.NewClient(cfg.WalletURL, log)
	ledgerClient := ledger.NewClient(cfg.LedgerURL, log)
	if !cfg.DevMode() && cfg.WalletURL == "" {
		log.Info("boot.wallet", "note", "WALLET_URL empty; wallet debit uses in-process stub")
	}
	deps := &app.Deps{
		Intents:   intents,
		Outbox:    outbox,
		Publisher: publisher,
		PSP:       router,
		Fraud:     fraudClient,
		Wallet:    walletClient,
		Ledger:    ledgerClient,
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
	}

	_, _ = deps.UpsertRoute(context.Background(), app.UpsertRouteInput{
		TenantID:   mustParse("11111111-1111-1111-1111-111111111111"),
		MethodType: domain.MethodCard, Currency: "TRY",
		Providers: []string{router.Name()}, Priority: 1,
	})

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				return err
			}
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
	if c, ok := any(publisher).(interface{ Close() error }); ok {
		_ = c.Close()
	}
	if db != nil {
		if err := db.Close(); err != nil {
			log.Error("postgres.close", "err", err)
		}
	}
	log.Info("shutdown.complete")
}

func mustParse(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
