// Package config loads checkout-service settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the runtime configuration for checkout-service.
type Config struct {
	HTTPAddr           string
	GRPCAddr           string
	PublicBaseURL      string
	DatabaseURL        string
	RedisURL           string
	KafkaBrokers       []string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
	MinOrderMinor      int64

	CartURL      string
	OrderURL     string
	PaymentURL   string
	InventoryURL string
	PricingURL   string
	PromoURL     string
	FraudURL     string
	GeofenceURL  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8087"),
		GRPCAddr:           env("GRPC_ADDR", ":9098"),
		PublicBaseURL:      strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:8087"), "/"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 240),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
		MinOrderMinor:      int64(envInt("MIN_ORDER_MINOR", 0)),
		CartURL:            env("CART_URL", ""),
		OrderURL:           env("ORDER_URL", ""),
		PaymentURL:         env("PAYMENT_URL", ""),
		InventoryURL:       env("INVENTORY_URL", ""),
		PricingURL:         env("PRICING_URL", ""),
		PromoURL:           strings.TrimRight(env("PROMO_URL", ""), "/"),
		FraudURL:           strings.TrimRight(firstNonEmpty(env("FRAUD_URL", ""), env("AI_PLATFORM_URL", "")), "/"),
		GeofenceURL:        strings.TrimRight(env("GEOFENCE_URL", ""), "/"),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	return cfg, nil
}

// DevMode is true when no DATABASE_URL is configured (in-memory repos).
func (c Config) DevMode() bool {
	return c.DatabaseURL == ""
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
