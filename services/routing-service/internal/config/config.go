package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the runtime configuration for routing-service.
type Config struct {
	HTTPAddr           string
	GRPCAddr           string
	PublicBaseURL      string
	DatabaseURL        string
	RedisURL           string
	KafkaBrokers       []string
	WeatherURL         string
	WeatherAPIKey      string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8097"),
		GRPCAddr:           env("GRPC_ADDR", ":9107"),
		PublicBaseURL:      strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:8097"), "/"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		WeatherURL:         firstEnv("WEATHER_URL", "OPENWEATHER_URL"),
		WeatherAPIKey:      firstEnv("WEATHER_API_KEY", "OPENWEATHER_API_KEY"),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 240),
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
