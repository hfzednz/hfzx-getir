// Package postgres implements production Inventory Service repositories against PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a *sql.DB using the pgx stdlib driver and verifies connectivity.
func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("postgres: empty DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}

// Repos aggregates postgres port implementations.
type Repos struct {
	Warehouses   *WarehouseRepo
	Locations    *LocationRepo
	Balances     *BalanceRepo
	Lots         *LotRepo
	Reservations *ReservationRepo
	Movements    *MovementRepo
	Transfers    *TransferRepo
	Counts       *CountRepo
	Returns      *ReturnRepo
	Forecasts    *ForecastRepo
}

// NewRepos wires all repository ports on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Warehouses:   &WarehouseRepo{DB: db},
		Locations:    &LocationRepo{DB: db},
		Balances:     &BalanceRepo{DB: db},
		Lots:         &LotRepo{DB: db},
		Reservations: &ReservationRepo{DB: db},
		Movements:    &MovementRepo{DB: db},
		Transfers:    &TransferRepo{DB: db},
		Counts:       &CountRepo{DB: db},
		Returns:      &ReturnRepo{DB: db},
		Forecasts:    &ForecastRepo{DB: db},
	}
}
