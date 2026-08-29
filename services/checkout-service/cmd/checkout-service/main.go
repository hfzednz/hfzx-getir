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
	"github.com/nexora/checkout-service/internal/adapters/cart"
	"github.com/nexora/checkout-service/internal/adapters/fraud"
	"github.com/nexora/checkout-service/internal/adapters/geofence"
	httpadapter "github.com/nexora/checkout-service/internal/adapters/http"
	"github.com/nexora/checkout-service/internal/adapters/inventory"
	"github.com/nexora/checkout-service/internal/adapters/kafka"
	"github.com/nexora/checkout-service/internal/adapters/order"
	"github.com/nexora/checkout-service/internal/adapters/payment"
	"github.com/nexora/checkout-service/internal/adapters/postgres"
	"github.com/nexora/checkout-service/internal/adapters/pricing"
	"github.com/nexora/checkout-service/internal/adapters/promo"
	"github.com/nexora/checkout-service/internal/adapters/redis"
	"github.com/nexora/checkout-service/internal/app"
	"github.com/nexora/checkout-service/internal/app/memory"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/config"
	"github.com/nexora/checkout-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	sessions, events, outbox := memory.NewRepos(store)
	rdb := redis.NewClient(cfg.RedisURL, log)
	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)

	// Demo cart for in-memory HTTP smoke tests (empty DATABASE_URL).
	demoTenant := mustParse("11111111-1111-1111-1111-111111111111")
	demoPrincipal := mustParse("22222222-2222-2222-2222-222222222222")
	demoCart := mustParse("33333333-3333-3333-3333-333333333333")
	demoVariant := mustParse("44444444-4444-4444-4444-444444444444")
	store.SeedCart(memoryCart(demoTenant, demoPrincipal, demoCart, demoVariant))

	memCart := &memory.CartClient{S: store}
	memPricing := &memory.PricingClient{}
	memInventory := memory.NewInventoryClient()
	memPay := &memory.PaymentEligibilityClient{}
	memOrders := &memory.OrderClient{}

	deps := &app.Deps{
		Sessions:      sessions,
		Events:        events,
		Outbox:        outbox,
		Publisher:     publisher,
		Cart:          memCart,
		Pricing:       memPricing,
		Inventory:     memInventory,
		Geofence:      &memory.GeofenceClient{},
		Fraud:         &memory.FraudClient{},
		PayElig:       memPay,
		Orders:        memOrders,
		Promo:         &memory.PromoClient{},
		Customer:      &memory.CustomerClient{},
		Clock:         app.SystemClock{},
		IDs:           app.UUIDGen{},
		CompleteLock:  &memory.CompleteLocker{S: store},
		MinOrderMinor: cfg.MinOrderMinor,
	}

	var db *sql.DB
	if !cfg.DevMode() {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		repos := postgres.NewRepos(db)
		deps.Sessions = repos.Sessions
		deps.Events = repos.Events
		deps.Outbox = repos.Outbox
		deps.Cart = cart.NewClient(cfg.CartURL, log, nil)
		deps.Orders = order.NewClient(cfg.OrderURL, log, nil)
		deps.PayElig = payment.NewClient(cfg.PaymentURL, log, nil)
		deps.Inventory = inventory.NewClient(cfg.InventoryURL, log, nil)
		deps.Pricing = pricing.NewClient(cfg.PricingURL, log, nil)
		deps.CompleteLock = &postgres.AdvisoryCompleteLocker{DB: db}
		log.Info("boot.database", "adapter", "postgres", "kafkaBrokers", len(cfg.KafkaBrokers))
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}

	// Always wrap so HTTP is used when URL is set (inner remains memory / prior deps).
	deps.Promo = promo.NewClient(cfg.PromoURL, log, deps.Promo)
	deps.Fraud = fraud.NewClient(cfg.FraudURL, log, deps.Fraud)
	deps.Geofence = geofence.NewClient(cfg.GeofenceURL, log, deps.Geofence)

	// Prefer HTTP cart/order when URLs are set (phone-test / staging), even in in-memory mode.
	deps.Cart = cart.NewClient(cfg.CartURL, log, deps.Cart)
	deps.Orders = order.NewClient(cfg.OrderURL, log, deps.Orders)

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
	if err := publisher.Close(); err != nil {
		log.Error("kafka.close", "err", err)
	}
	if err := rdb.Close(); err != nil {
		log.Error("redis.close", "err", err)
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

func memoryCart(tenant, principal, cartID, variant uuid.UUID) ports.CartView {
	return ports.CartView{
		ID: cartID, TenantID: tenant, PrincipalID: principal,
		CityID: "istanbul", Currency: "TRY", Active: true,
		Lines: []ports.CartLine{{
			VariantID: variant, SKUCode: "SKU-1", TitleSnapshot: "Milk",
			Qty: 2, UnitPriceMinor: 1500,
		}},
	}
}
