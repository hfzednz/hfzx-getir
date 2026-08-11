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

	grpcadapter "github.com/nexora/security-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/security-service/internal/adapters/http"
	"github.com/nexora/security-service/internal/adapters/kafka"
	"github.com/nexora/security-service/internal/adapters/postgres"
	"github.com/nexora/security-service/internal/app"
	"github.com/nexora/security-service/internal/app/memory"
	"github.com/nexora/security-service/internal/config"
	"github.com/nexora/security-service/internal/ratelimit"
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
		Policies: repos.Policies, Audits: repos.Audits, Secrets: repos.Secrets,
		Threats: repos.Threats, Vulns: repos.Vulns, Incidents: repos.Incidents,
		Compliance: repos.Compliance, DataGov: repos.DataGov, Risks: repos.Risks,
		Access: repos.Access, Devices: repos.Devices, AISec: repos.AISec,
		FraudSigs: repos.FraudSigs, Outbox: repos.Outbox,
		Vault: repos.Vault, OPA: repos.OPA, Identity: repos.Identity,
		Fraud: repos.Fraud, SIEM: repos.SIEM, SOAR: repos.SOAR, AIGuard: repos.AIGuard,
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{}, IDs: app.UUIDGen{},
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
		deps.Policies = pg.Policies
		deps.Audits = pg.Audits
		deps.Secrets = pg.Secrets
		deps.Threats = pg.Threats
		deps.Vulns = pg.Vulns
		deps.Incidents = pg.Incidents
		deps.Compliance = pg.Compliance
		deps.DataGov = pg.DataGov
		deps.Risks = pg.Risks
		deps.Access = pg.Access
		deps.Devices = pg.Devices
		deps.AISec = pg.AISec
		deps.FraudSigs = pg.FraudSigs
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	}
	_ = db
	if cfg.VaultAddr != "" {
		log.Info("boot.vault", "addr", cfg.VaultAddr)
	}
	if cfg.OPAURL != "" {
		log.Info("boot.opa", "url", cfg.OPAURL)
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()
	srv := httpadapter.NewServer(httpadapter.ServerConfig{
		Addr: cfg.HTTPAddr, Deps: deps, Limiter: ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute, CORSOrigins: cfg.CORSAllowedOrigins, Log: log,
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
	_ = srv.Shutdown(ctx)
	log.Info("shutdown.complete")
}
