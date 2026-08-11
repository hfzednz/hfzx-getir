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
		return nil, fmt.Errorf("DATABASE_URL required")
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

// Repos aggregates all postgres port implementations for location-service.
type Repos struct {
	Addresses *AddressRepo
	POIs      *POIRepo
	History   *HistoryRepo
	Cache     *CacheRepo
	Heat      *HeatRepo
	Outbox    *OutboxRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Addresses: &AddressRepo{DB: db},
		POIs:      &POIRepo{DB: db},
		History:   &HistoryRepo{DB: db},
		Cache:     &CacheRepo{DB: db},
		Heat:      &HeatRepo{DB: db},
		Outbox:    &OutboxRepo{DB: db},
	}
}
