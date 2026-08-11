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
	KafkaBrokers       []string
	VaultAddr          string
	OPAURL             string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8109"),
		GRPCAddr:           env("GRPC_ADDR", ":9109"),
		DatabaseURL:        env("DATABASE_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		VaultAddr:          env("VAULT_ADDR", ""),
		OPAURL:             env("OPA_URL", ""),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 1200),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	return cfg, nil
}

func (c Config) DevMode() bool { return c.DatabaseURL == "" }

func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}

func envInt(k string, f int) int {
	v := os.Getenv(k)
	if v == "" {
		return f
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return f
	}
	return n
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
