package httpclients

import (
	"os"

	"github.com/nexora/bff-customer/internal/app"
)

// Config holds downstream base URLs (defaults match docs/launch/service-registry.yaml).
type Config struct {
	IdentityURL       string
	CatalogURL        string
	CartURL           string
	CheckoutURL       string
	OrderURL          string
	TrackingURL       string
	PaymentURL        string
	RecommendationURL string
	LocationURL       string
	NotificationURL   string
	CRMURL            string
	ReviewURL         string
	InventoryURL      string
	PromoURL          string
}

// ConfigFromEnv loads URLs from environment with localhost registry defaults.
func ConfigFromEnv() Config {
	return Config{
		IdentityURL:       envOr("IDENTITY_URL", "http://127.0.0.1:8081"),
		CatalogURL:        envOr("CATALOG_URL", "http://127.0.0.1:8083"),
		CartURL:           envOr("CART_URL", "http://127.0.0.1:8086"),
		CheckoutURL:       envOr("CHECKOUT_URL", "http://127.0.0.1:8087"),
		OrderURL:          envOr("ORDER_URL", "http://127.0.0.1:8085"),
		TrackingURL:       envOr("TRACKING_URL", "http://127.0.0.1:8098"),
		PaymentURL:        envOr("PAYMENT_URL", "http://127.0.0.1:8089"),
		RecommendationURL: envOr("RECOMMENDATION_URL", "http://127.0.0.1:8105"),
		LocationURL:       envOr("LOCATION_URL", "http://127.0.0.1:8100"),
		NotificationURL:   envOr("NOTIFICATION_URL", "http://127.0.0.1:8101"),
		CRMURL:            envOr("CRM_URL", "http://127.0.0.1:8102"),
		ReviewURL:         envOr("REVIEW_URL", "http://127.0.0.1:8103"),
		InventoryURL:      envOr("INVENTORY_URL", "http://127.0.0.1:8084"),
		PromoURL:          envOr("PROMO_URL", "http://127.0.0.1:8094"),
	}
}

func envOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

// NewDeps wires production HTTP clients for all BFF ports.
func NewDeps(cfg Config) *app.Deps {
	return &app.Deps{
		Identity: NewIdentity(cfg.IdentityURL),
		Catalog:  NewCatalog(cfg.CatalogURL),
		Recs:     NewRecs(cfg.RecommendationURL),
		Cart:     NewCart(cfg.CartURL),
		Checkout: NewCheckout(cfg.CheckoutURL, cfg.PaymentURL),
		Orders:   NewOrders(cfg.OrderURL),
		Tracking: NewTracking(cfg.TrackingURL),
		Location: NewLocation(cfg.LocationURL),
		Notify:   NewNotify(cfg.NotificationURL),
		CRM:      NewCRM(cfg.CRMURL),
		Reviews:  NewReviews(cfg.ReviewURL),
		Stores:   NewStores(cfg.InventoryURL),
		Promo:    NewPromo(cfg.PromoURL),
	}
}
