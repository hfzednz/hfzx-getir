// Package config loads customer-profile-service settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration for customer-profile-service.
type Config struct {
	HTTPAddr string
	GRPCAddr string

	DatabaseURL  string
	RedisURL     string
	KafkaBrokers []string
	SearchURL    string

	RateLimitPerMinute int
	CORSAllowedOrigins []string

	// MediaServiceURL is an optional media-service base URL (avatar port).
	MediaServiceURL string
	// GeofenceURL is an optional geofence-service base URL (zone port).
	GeofenceURL string

	ShutdownTimeout time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8082"),
		GRPCAddr:           env("GRPC_ADDR", ":9091"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		SearchURL:          firstEnv("SEARCH_URL", "OPENSEARCH_URL"),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 120),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
		MediaServiceURL:    env("MEDIA_SERVICE_URL", ""),
		GeofenceURL:        env("GEOFENCE_URL", ""),
		ShutdownTimeout:    envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
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

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
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
