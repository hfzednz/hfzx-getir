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

	grpcadapter "github.com/nexora/loyalty-service/internal/adapters/grpc"
	httpadapter "github.com/nexora/loyalty-service/internal/adapters/http"
	"github.com/nexora/loyalty-service/internal/adapters/kafka"
	"github.com/nexora/loyalty-service/internal/adapters/postgres"
	"github.com/nexora/loyalty-service/internal/adapters/redis"
	walletadapter "github.com/nexora/loyalty-service/internal/adapters/wallet"
	"github.com/nexora/loyalty-service/internal/app"
	"github.com/nexora/loyalty-service/internal/app/memory"
	"github.com/nexora/loyalty-service/internal/config"
	"github.com/nexora/loyalty-service/internal/ratelimit"
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
		Accounts: repos.Accounts, Memberships: repos.Memberships, Rewards: repos.Rewards,
		Referrals: repos.Referrals, Missions: repos.Missions, Achievements: repos.Achievements,
		Streaks: repos.Streaks, Spins: repos.Spins, Collectibles: repos.Collectibles,
		Cashbacks: repos.Cashbacks, AIScores: repos.AIScores, Outbox: repos.Outbox,
		Wallet:    walletadapter.NewClient(cfg.WalletBaseURL, log),
		Publisher: kafka.NewPublisher(cfg.KafkaBrokers, log),
		Clock:     app.SystemClock{},
		IDs:       app.UUIDGen{},
		Rand:      app.NewSystemRand(time.Now().UnixNano()),
	}

	var db *sql.DB
	if !cfg.DevMode() {
		var openErr error
		db, openErr = postgres.Open(cfg.DatabaseURL)
		if openErr != nil {
			log.Error("postgres.open", "err", openErr)
			os.Exit(1)
		}
		pg := postgres.NewRepos(db)
		deps.Accounts = pg.Accounts
		deps.Memberships = pg.Memberships
		deps.Rewards = pg.Rewards
		deps.Referrals = pg.Referrals
		deps.Missions = pg.Missions
		deps.Achievements = pg.Achievements
		deps.Streaks = pg.Streaks
		deps.Spins = pg.Spins
		deps.Collectibles = pg.Collectibles
		deps.Cashbacks = pg.Cashbacks
		deps.AIScores = pg.AIScores
		deps.Outbox = pg.Outbox
		log.Info("boot.database", "driver", "pgx", "repos", "postgres")
	} else {
		log.Info("boot.dev_mode", "reason", "DATABASE_URL empty; using in-memory repositories")
	}
	if cfg.RedisURL != "" {
		if _, err := redis.Open(cfg.RedisURL); err != nil {
			log.Warn("redis.open", "err", err)
		}
	}

	grpcadapter.NewServer(cfg.GRPCAddr, log).Start()

	ready := func(*http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return db.PingContext(ctx)
		}
		return nil
	}

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
	log.Info("shutdown.complete")
}
