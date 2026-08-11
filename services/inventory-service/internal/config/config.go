// Package config loads inventory-service settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration for inventory-service.
type Config struct {
	HTTPAddr           string
	GRPCAddr           string
	PublicBaseURL      string
	DatabaseURL        string
	RedisURL           string
	KafkaBrokers       []string
	OpenSearchURL      string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
	SoftReserveTTL     time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8083"),
		GRPCAddr:           env("GRPC_ADDR", ":9093"),
		PublicBaseURL:      strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:8083"), "/"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		OpenSearchURL:      env("OPENSEARCH_URL", ""),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 240),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
		SoftReserveTTL:     envDuration("SOFT_RESERVE_TTL", 15*time.Minute),
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

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
