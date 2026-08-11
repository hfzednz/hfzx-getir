package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr           string
	GRPCAddr           string
	DatabaseURL        string
	RedisURL           string
	KafkaBrokers       []string
	OpenSearchURL      string
	VectorURL          string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8104"),
		GRPCAddr:           env("GRPC_ADDR", ":9104"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		OpenSearchURL:      env("OPENSEARCH_URL", ""),
		VectorURL:          env("VECTOR_URL", ""),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 600),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	return cfg, nil
}

func (c Config) DevMode() bool { return c.DatabaseURL == "" }

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
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
