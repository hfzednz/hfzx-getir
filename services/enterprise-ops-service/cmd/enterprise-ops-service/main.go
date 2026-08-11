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

	grpcadapter "github.com/nexora/enterprise-ops-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/enterprise-ops-service/internal/adapters/http"
	"github.com/nexora/enterprise-ops-service/internal/adapters/kafka"
	"github.com/nexora/enterprise-ops-service/internal/adapters/postgres"
	"github.com/nexora/enterprise-ops-service/internal/app"
	"github.com/nexora/enterprise-ops-service/internal/app/memory"
	"github.com/nexora/enterprise-ops-service/internal/config"
	"github.com/nexora/enterprise-ops-service/internal/ratelimit"
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
		Org: repos.OrgR, Policies: repos.PolicyR, Portfolios: repos.PortfolioR,
		Programs: repos.ProgramR, Projects: repos.ProjectR, Milestones: repos.MilestoneR,
		Objectives: repos.ObjectiveR, KeyResults: repos.KeyResultR, KPIs: repos.KPIR,
		Risks: repos.RiskR, Continuity: repos.ContinuityR, Audits: repos.AuditR,
		Findings: repos.FindingR, Meetings: repos.MeetingR, Decisions: repos.DecisionR,
		Knowledge: repos.KnowledgeR, Resources: repos.ResourceR, Outbox: repos.OutboxR,
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Security: repos.Security, AI: repos.AI, Metrics: repos.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
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
		deps.Org = pg.Org
		deps.Policies = pg.Policies
		deps.Portfolios = pg.Portfolios
		deps.Programs = pg.Programs
		deps.Projects = pg.Projects
		deps.Milestones = pg.Milestones
		deps.Objectives = pg.Objectives
		deps.KeyResults = pg.KeyResults
		deps.KPIs = pg.KPIs
		deps.Risks = pg.Risks
		deps.Continuity = pg.Continuity
		deps.Audits = pg.Audits
		deps.Findings = pg.Findings
		deps.Meetings = pg.Meetings
		deps.Decisions = pg.Decisions
		deps.Knowledge = pg.Knowledge
		deps.Resources = pg.Resources
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
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
	if db != nil {
		_ = db.Close()
	}
	log.Info("shutdown.complete")
}
