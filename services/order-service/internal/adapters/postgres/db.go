// Package postgres implements production Order Service repositories against PostgreSQL.
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

// Repos aggregates postgres port implementations.
type Repos struct {
	Orders       *OrderRepo
	Events       *EventStoreRepo
	Sagas        *SagaRepo
	Outbox       *OutboxRepo
	Fulfillments *FulfillmentRepo
	Returns      *ReturnRepo
	Refunds      *RefundRepo
}

// NewRepos wires all repository ports on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Orders:       &OrderRepo{DB: db},
		Events:       &EventStoreRepo{DB: db},
		Sagas:        &SagaRepo{DB: db},
		Outbox:       &OutboxRepo{DB: db},
		Fulfillments: &FulfillmentRepo{DB: db},
		Returns:      &ReturnRepo{DB: db},
		Refunds:      &RefundRepo{DB: db},
	}
}
