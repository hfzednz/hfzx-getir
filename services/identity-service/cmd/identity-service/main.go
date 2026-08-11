package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	httpadapter "github.com/nexora/identity-service/internal/adapters/http"
	"github.com/nexora/identity-service/internal/adapters/http/oidc"
	"github.com/nexora/identity-service/internal/adapters/kafka"
	"github.com/nexora/identity-service/internal/adapters/postgres"
	redisadapter "github.com/nexora/identity-service/internal/adapters/redis"
	"github.com/nexora/identity-service/internal/adapters/sms"
	"github.com/nexora/identity-service/internal/adapters/social"
	"github.com/nexora/identity-service/internal/app"
	"github.com/nexora/identity-service/internal/app/memory"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/config"
	"github.com/nexora/identity-service/internal/ratelimit"
	"github.com/nexora/identity-service/internal/security/jwt"
	"github.com/nexora/identity-service/internal/security/password"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	keys := jwt.NewKeyManager()
	if cfg.JWTKeyPEM != "" {
		if err := keys.LoadPEM(cfg.JWTKeyPEM, ""); err != nil {
			log.Error("jwt.load_pem", "err", err)
			os.Exit(1)
		}
	} else if cfg.DevMode() {
		if err := keys.Generate(2048); err != nil {
			log.Error("jwt.generate", "err", err)
			os.Exit(1)
		}
		log.Info("jwt.ephemeral_key", "kid", keys.KID())
	} else {
		log.Error("jwt.required", "err", "JWT_KEY_PEM required when DATABASE_URL is set")
		os.Exit(1)
	}

	publisher := kafka.NewPublisher(cfg.KafkaBrokers, log)
	if cfg.DevMode() {
		publisher.AllowNoopWithoutBrokers = true
	} else if len(cfg.KafkaBrokers) == 0 {
		log.Error("boot.kafka", "err", "KAFKA_BROKERS required when DATABASE_URL is set")
		os.Exit(1)
	}

	otpSender, err := buildOTPSender(cfg, log)
	if err != nil {
		log.Error("otp.sender", "err", err)
		os.Exit(1)
	}

	wa := webauthn.NewService(webauthn.Config{
		RPDisplayName: "NEXORA",
		RPID:          envOr("WEBAUTHN_RP_ID", "localhost"),
		RPOrigins:     splitOrigins(envOr("WEBAUTHN_ORIGINS", cfg.PublicBaseURL)),
		Timeout:       60 * time.Second,
	})

	otpPepper := strings.TrimSpace(os.Getenv("OTP_PEPPER"))
	if otpPepper == "" {
		if cfg.DevMode() {
			otpPepper = "nexora-otp-pepper"
		} else {
			log.Error("otp.pepper", "err", "OTP_PEPPER required when DATABASE_URL is set")
			os.Exit(1)
		}
	} else if !cfg.DevMode() && otpPepper == "nexora-otp-pepper" {
		log.Error("otp.pepper", "err", "OTP_PEPPER must not use the known default in production")
		os.Exit(1)
	}

	deps := &app.Deps{
		OTP:       otpSender,
		Events:    publisher,
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
		Passwords: password.NewDefaultHasher(),
		WebAuthn:  wa,
		JWTKeys:   keys,
		Issuer:    cfg.JWTIssuer,
		Audience:  cfg.JWTAudience,
		AccessTTL: cfg.AccessTTL,
		OTPPepper: otpPepper,
		Social:    social.LoadProvidersFromEnv(),
	}
	deps.Tokens = &app.DefaultTokenIssuer{Deps: deps}

	var dbReady func(*http.Request) error
	var dbCloser interface{ Close() error }
	if cfg.DevMode() {
		store := memory.NewStore()
		deps.Principals = &memory.PrincipalRepo{S: store}
		deps.Sessions = &memory.SessionRepo{S: store}
		deps.Devices = &memory.DeviceRepo{S: store}
		deps.Roles = &memory.RoleRepo{S: store}
		deps.Audit = &memory.AuditRepo{S: store}
		deps.OAuth = &memory.OAuthRepo{S: store}
		deps.Risk = &memory.RiskRepo{S: store}
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
		dbReady = func(*http.Request) error { return nil }
	} else {
		db, err := postgres.Open(cfg.DatabaseURL)
		if err != nil {
			log.Error("postgres.open", "err", err)
			os.Exit(1)
		}
		dbCloser = db
		deps.Principals = &postgres.PrincipalRepo{DB: db}
		deps.Sessions = &postgres.SessionRepo{DB: db}
		deps.Devices = &postgres.DeviceRepo{DB: db}
		deps.Roles = &postgres.RoleRepo{DB: db}
		deps.Audit = &postgres.AuditRepo{DB: db}
		deps.OAuth = &postgres.OAuthRepo{DB: db}
		deps.Risk = &postgres.RiskRepo{DB: db}
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
		dbReady = func(*http.Request) error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		}
	}

	var sessionCache *redisadapter.SessionCache
	if cfg.RedisURL != "" {
		sc, err := redisadapter.NewSessionCache(cfg.RedisURL)
		if err != nil {
			log.Warn("boot.redis", "err", err)
		} else {
			sessionCache = sc
			log.Info("boot.redis", "adapter", "session-cache")
		}
	}

	ready := func(r *http.Request) error {
		if err := dbReady(r); err != nil {
			return err
		}
		if sessionCache != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return sessionCache.Ping(ctx)
		}
		return nil
	}

	oidcProvider := oidc.NewProvider(oidc.Config{
		Issuer:    cfg.OIDCIssuer,
		Audience:  cfg.JWTAudience,
		Keys:      keys,
		AccessTTL: cfg.AccessTTL,
	})

	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr:               cfg.HTTPAddr,
		Deps:               deps,
		OIDC:               oidcProvider,
		Limiter:            ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute,
		CORSOrigins:        cfg.CORSAllowedOrigins,
		Log:                log,
		Live:               func(*http.Request) error { return nil },
		Ready:              ready,
	})

	go func() {
		log.Info("http.listen", "addr", cfg.HTTPAddr, "devMode", cfg.DevMode(), "otpDevMode", cfg.OTPDevMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http.serve", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = publisher.Close()
	if sessionCache != nil {
		if err := sessionCache.Close(); err != nil {
			log.Error("redis.close", "err", err)
		}
	}
	if dbCloser != nil {
		_ = dbCloser.Close()
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	log.Info("shutdown.complete")
}

func buildOTPSender(cfg config.Config, log *slog.Logger) (ports.OTPSender, error) {
	if cfg.OTPDevMode {
		return &loggingOTPSender{log: log, enabled: true}, nil
	}
	if os.Getenv("TWILIO_ACCOUNT_SID") != "" {
		return sms.NewTwilioFromEnv()
	}
	if os.Getenv("SMS_WEBHOOK_URL") != "" {
		return sms.NewWebhookFromEnv()
	}
	return nil, errOTPProviderRequired{}
}

type errOTPProviderRequired struct{}

func (errOTPProviderRequired) Error() string {
	return "OTP_DEV_MODE=false requires TWILIO_* or SMS_WEBHOOK_URL"
}

type loggingOTPSender struct {
	log     *slog.Logger
	enabled bool
}

func (s *loggingOTPSender) SendOTP(_ context.Context, tenantID uuid.UUID, phone, code string) error {
	if s.enabled {
		s.log.Info("otp.dev_mode", "phone", phone, "code", code, "tenantId", tenantID.String())
	}
	return nil
}

func envOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func splitOrigins(s string) []string {
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
