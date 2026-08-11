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

	grpcadapter "github.com/nexora/ai-platform-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/ai-platform-service/internal/adapters/http"
	"github.com/nexora/ai-platform-service/internal/adapters/kafka"
	"github.com/nexora/ai-platform-service/internal/adapters/postgres"
	"github.com/nexora/ai-platform-service/internal/adapters/python_runtime"
	"github.com/nexora/ai-platform-service/internal/app"
	"github.com/nexora/ai-platform-service/internal/app/memory"
	"github.com/nexora/ai-platform-service/internal/config"
	"github.com/nexora/ai-platform-service/internal/ratelimit"
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

	runtime := python_runtime.NewClient(cfg.PythonRuntimeURL, memory.HeuristicRuntime{})

	deps := &app.Deps{
		Features: repos.Features, Models: repos.Models, Prompts: repos.Prompts,
		Memory: repos.Memory, Agents: repos.Agents, Automation: repos.Automation,
		Drift: repos.Drift, Outbox: repos.Outbox,
		Runtime: runtime, LLM: memory.MockLLM{}, RAG: memory.MockRAG{}, Embed: memory.MockEmbed{},
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
		deps.Features = pg.Features
		deps.Models = pg.Models
		deps.Prompts = pg.Prompts
		deps.Memory = pg.Memory
		deps.Agents = pg.Agents
		deps.Automation = pg.Automation
		deps.Drift = pg.Drift
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
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
		Addr: cfg.HTTPAddr, Deps: deps, Limiter: ratelimit.NewMemoryLimiter(),
		RateLimitPerMinute: cfg.RateLimitPerMinute, CORSOrigins: cfg.CORSAllowedOrigins, Log: log,
		Live:  func(*http.Request) error { return nil },
		Ready: ready,
	})

	go func() {
		log.Info("http.listen", "addr", cfg.HTTPAddr, "devMode", cfg.DevMode(), "pythonRuntime", cfg.PythonRuntimeURL)
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
