package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcadapter "github.com/nexora/notification-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/notification-service/internal/adapters/http"
	"github.com/nexora/notification-service/internal/adapters/kafka"
	"github.com/nexora/notification-service/internal/adapters/postgres"
	"github.com/nexora/notification-service/internal/adapters/providers"
	"github.com/nexora/notification-service/internal/adapters/redis"
	"github.com/nexora/notification-service/internal/app"
	"github.com/nexora/notification-service/internal/app/memory"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/config"
	"github.com/nexora/notification-service/internal/ratelimit"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config.load", "err", err)
		os.Exit(1)
	}

	store := memory.NewStore()
	repos := memory.NewRepos(store)

	deps := &app.Deps{
		Templates: repos.Templates, Messages: repos.Messages, Preferences: repos.Preferences,
		Devices: repos.Devices, Inbox: repos.Inbox, Schedules: repos.Schedules,
		Deliveries: repos.Deliveries, Outbox: repos.Outbox,
		Push: buildPush(cfg, log), Email: buildEmail(cfg, log),
		SMS: buildSMS(cfg, log), WhatsApp: buildWhatsApp(cfg, log),
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
	}

	var db *sql.DB
	if cfg.DevMode() {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	} else {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		pg := postgres.NewRepos(db)
		deps.Templates = pg.Templates
		deps.Messages = pg.Messages
		deps.Preferences = pg.Preferences
		deps.Devices = pg.Devices
		deps.Inbox = pg.Inbox
		deps.Schedules = pg.Schedules
		deps.Deliveries = pg.Deliveries
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}
	if cfg.RedisURL != "" {
		if _, err := redis.Open(cfg.RedisURL); err != nil {
			log.Warn("redis.open", "err", err)
		}
	}

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		}
		return nil
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()

	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr:               cfg.HTTPAddr,
		Deps:               deps,
		Limiter:            ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute,
		CORSOrigins:        cfg.CORSAllowedOrigins,
		Log:                log,
		Live:               func(*http.Request) error { return nil },
		Ready:              ready,
	})

	go func() {
		log.Info("http.listen", "addr", cfg.HTTPAddr, "devMode", cfg.DevMode())
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
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("http.shutdown", "err", err)
		os.Exit(1)
	}
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}

func buildEmail(cfg config.Config, log *slog.Logger) ports.EmailProvider {
	if e, err := providers.NewSMTPFromEnv(); err == nil {
		log.Info("email.provider", "driver", "smtp")
		return e
	} else if !cfg.DevMode() {
		log.Error("email.required", "err", err)
		os.Exit(1)
	}
	return &providers.MockEmail{Log: log}
}

func buildSMS(cfg config.Config, log *slog.Logger) ports.SMSProvider {
	if s, err := providers.NewTwilioSMSFromEnv(); err == nil {
		log.Info("sms.provider", "driver", "twilio")
		return s
	} else if !cfg.DevMode() {
		log.Error("sms.required", "err", err)
		os.Exit(1)
	}
	return &providers.MockSMS{Log: log}
}

func buildPush(cfg config.Config, log *slog.Logger) ports.PushProvider {
	if p, err := providers.NewFCMFromEnv(); err == nil {
		log.Info("push.provider", "driver", "fcm")
		return p
	} else if !cfg.DevMode() {
		log.Error("push.required", "err", err)
		os.Exit(1)
	}
	return &providers.MockPush{Log: log}
}

func buildWhatsApp(cfg config.Config, log *slog.Logger) ports.WhatsAppProvider {
	if w, err := providers.NewWhatsAppFromEnv(); err == nil {
		log.Info("whatsapp.provider", "driver", "meta_cloud")
		return w
	} else if !cfg.DevMode() {
		log.Warn("whatsapp.optional", "err", err)
	}
	return &providers.MockWhatsApp{Log: log}
}
