// Package postgres implements production wallet-service repositories against PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a *sql.DB using the pgx stdlib driver and verifies connectivity.
func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty database url")
	}
	db, err := sql.Open("pgx", dsn)
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

// Repos groups wallet persistence adapters.
type Repos struct {
	Wallets *WalletRepo
	Outbox  *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Wallets: &WalletRepo{DB: db},
		Outbox:  &OutboxRepo{DB: db},
	}
}
