// Package postgres implements production warehouse-service repositories against PostgreSQL.
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

// Repos aggregates all postgres port implementations for warehouse-service.
type Repos struct {
	Fulfillments *FulfillmentRepo
	Tasks        *TaskRepo
	Picks        *PickRepo
	Packs        *PackRepo
	Dispatches   *DispatchRepo
	Stations     *StationRepo
	Workforce    *WorkforceRepo
	Equipment    *EquipmentRepo
	QC           *QCRepo
	Labels       *LabelRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Fulfillments: &FulfillmentRepo{DB: db},
		Tasks:        &TaskRepo{DB: db},
		Picks:        &PickRepo{DB: db},
		Packs:        &PackRepo{DB: db},
		Dispatches:   &DispatchRepo{DB: db},
		Stations:     &StationRepo{DB: db},
		Workforce:    &WorkforceRepo{DB: db},
		Equipment:    &EquipmentRepo{DB: db},
		QC:           &QCRepo{DB: db},
		Labels:       &LabelRepo{DB: db},
	}
}
