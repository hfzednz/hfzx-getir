package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a *sql.DB using the pgx stdlib driver and verifies connectivity.
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("empty database url")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return db, nil
}

// Repos aggregates all postgres port implementations for review-service.
type Repos struct {
	Reviews      *ReviewRepo
	Ratings      *RatingRepo
	Media        *MediaRepo
	Interactions *InteractionRepo
	Quality      *QualityRepo
	Moderation   *ModerationRepo
	Trust        *TrustRepo
	Reputation   *ReputationRepo
	Outbox       *OutboxRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Reviews:      &ReviewRepo{DB: db},
		Ratings:      &RatingRepo{DB: db},
		Media:        &MediaRepo{DB: db},
		Interactions: &InteractionRepo{DB: db},
		Quality:      &QualityRepo{DB: db},
		Moderation:   &ModerationRepo{DB: db},
		Trust:        &TrustRepo{DB: db},
		Reputation:   &ReputationRepo{DB: db},
		Outbox:       &OutboxRepo{DB: db},
	}
}
