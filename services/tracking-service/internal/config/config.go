package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the runtime configuration for tracking-service.
type Config struct {
	HTTPAddr           string
	GRPCAddr           string
	PublicBaseURL      string
	DatabaseURL        string
	RedisURL           string
	KafkaBrokers       []string
	GeofenceBaseURL    string
	HistoryCap         int
	ArrivalThresholdM  float64
	RateLimitPerMinute int
	CORSAllowedOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8098"),
		GRPCAddr:           env("GRPC_ADDR", ":9108"),
		PublicBaseURL:      strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:8098"), "/"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		GeofenceBaseURL:    env("GEOFENCE_BASE_URL", "http://localhost:8099"),
		HistoryCap:         envInt("HISTORY_CAP", 100),
		ArrivalThresholdM:  envFloat("ARRIVAL_THRESHOLD_M", 50),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 600),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	return cfg, nil
}

// DevMode is true when no DATABASE_URL is configured.
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

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
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
