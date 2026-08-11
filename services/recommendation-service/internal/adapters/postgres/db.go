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

// Repos aggregates all postgres port implementations for recommendation-service.
type Repos struct {
	Features *FeatureRepo
	Signals  *SignalRepo
	CoOccur  *CoOccurRepo
	Outbox   *OutboxRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Features: &FeatureRepo{DB: db},
		Signals:  &SignalRepo{DB: db},
		CoOccur:  &CoOccurRepo{DB: db},
		Outbox:   &OutboxRepo{DB: db},
	}
}
