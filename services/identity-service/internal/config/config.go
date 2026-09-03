// Package config loads identity-service settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration for identity-service.
type Config struct {
	HTTPAddr     string
	GRPCAddr     string
	PublicBaseURL string

	DatabaseURL string
	RedisURL    string
	KafkaBrokers []string

	JWTIssuer   string
	JWTAudience string
	JWTKeyPEM   string // optional path to RSA private key PEM
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	SessionIdleTTL time.Duration
	SessionAbsoluteTTL time.Duration

	OTPDevMode bool // log OTP codes instead of sending
	OTPTTL     time.Duration
	OTPLength  int

	RateLimitPerMinute int
	CORSAllowedOrigins []string

	OIDCIssuer string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8081"),
		GRPCAddr:           env("GRPC_ADDR", ":9090"),
		PublicBaseURL:      strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:8081"), "/"),
		DatabaseURL:        env("DATABASE_URL", ""),
		RedisURL:           env("REDIS_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		JWTIssuer:          env("JWT_ISS", "https://identity.nexora.local"),
		JWTAudience:        env("JWT_AUD", "nexora"),
		JWTKeyPEM:          env("JWT_KEY_PEM", ""),
		AccessTTL:          envDuration("ACCESS_TTL", 15*time.Minute),
		RefreshTTL:         envDuration("REFRESH_TTL", 30*24*time.Hour),
		SessionIdleTTL:     envDuration("SESSION_IDLE_TTL", 30*time.Minute),
		SessionAbsoluteTTL: envDuration("SESSION_ABSOLUTE_TTL", 12*time.Hour),
		OTPDevMode:         envBool("OTP_DEV_MODE", env("DATABASE_URL", "") == ""),
		OTPTTL:             envDuration("OTP_TTL", 5*time.Minute),
		OTPLength:          envInt("OTP_LENGTH", 6),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 120),
		CORSAllowedOrigins: splitCSV(env("CORS_ALLOWED_ORIGINS", "*")),
	}
	cfg.OIDCIssuer = env("OIDC_ISSUER", cfg.PublicBaseURL)
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	if cfg.JWTIssuer == "" || cfg.JWTAudience == "" {
		return Config{}, fmt.Errorf("JWT_ISS and JWT_AUD are required")
	}
	if !cfg.DevMode() {
		if cfg.JWTKeyPEM == "" {
			return Config{}, fmt.Errorf("JWT_KEY_PEM is required when DATABASE_URL is set")
		}
		if cfg.OTPDevMode {
			return Config{}, fmt.Errorf("OTP_DEV_MODE must be false when DATABASE_URL is set")
		}
		origins := cfg.CORSAllowedOrigins
		if len(origins) == 1 && origins[0] == "*" {
			return Config{}, fmt.Errorf("CORS_ALLOWED_ORIGINS must not be * when DATABASE_URL is set")
		}
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

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
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
