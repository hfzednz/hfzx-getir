package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpadapter "github.com/nexora/bff-admin/internal/adapters/http"
	"github.com/nexora/bff-admin/internal/adapters/httpclients"
	"github.com/nexora/bff-admin/internal/app"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8114"
	}
	cfg := httpclients.ConfigFromEnv()
	deps := &app.Deps{
		Orders:     httpclients.OrderClient{Client: httpclients.New(cfg.OrderURL)},
		LiveOps:    httpclients.LiveOpsClient{Client: httpclients.New(cfg.LiveOpsURL)},
		Catalog:    httpclients.CatalogClient{Client: httpclients.New(cfg.CatalogURL)},
		CRM:        httpclients.CRMClient{Client: httpclients.New(cfg.CrmURL)},
		Ledger:     httpclients.LedgerClient{Client: httpclients.NewLedger(cfg.LedgerURL)},
		Promo:      httpclients.PromoClient{Client: httpclients.New(cfg.PromoURL)},
		Pricing:    httpclients.PricingClient{Client: httpclients.New(cfg.PricingURL)},
		Inventory:  httpclients.InventoryClient{Client: httpclients.New(cfg.InventoryURL)},
		Identity:   httpclients.IdentityClient{Client: httpclients.New(cfg.IdentityURL)},
		Profile:    httpclients.ProfileClient{Client: httpclients.New(cfg.ProfileURL)},
		Settlement: httpclients.SettlementClient{Client: httpclients.New(cfg.SettlementURL)},
		Notify:     httpclients.NotifyClient{Client: httpclients.New(cfg.NotifyURL)},
		Loyalty:    httpclients.LoyaltyClient{Client: httpclients.New(cfg.LoyaltyURL)},
		AI:         httpclients.AIClient{Client: httpclients.New(cfg.AIURL)},
		Tracking:   httpclients.TrackingClient{Client: httpclients.New(cfg.TrackingURL)},
		Health: []app.HealthTarget{
			{Name: "order-service", URL: cfg.OrderURL},
			{Name: "liveops-service", URL: cfg.LiveOpsURL},
			{Name: "catalog-service", URL: cfg.CatalogURL},
			{Name: "crm-service", URL: cfg.CrmURL},
			{Name: "ledger-service", URL: cfg.LedgerURL},
			{Name: "promo-service", URL: cfg.PromoURL},
			{Name: "inventory-service", URL: cfg.InventoryURL},
			{Name: "identity-service", URL: cfg.IdentityURL},
			{Name: "settlement-service", URL: cfg.SettlementURL},
			{Name: "notification-service", URL: cfg.NotifyURL},
			{Name: "tracking-service", URL: cfg.TrackingURL},
		},
	}
	srv := httpadapter.NewServer(addr, deps)
	go func() {
		log.Println("listen", addr, "order", cfg.OrderURL, "catalog", cfg.CatalogURL, "crm", cfg.CrmURL, "ledger", cfg.LedgerURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Close()
}
