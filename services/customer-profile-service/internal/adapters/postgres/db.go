// Package postgres implements production customer-profile-service repositories against PostgreSQL.
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
		return nil, fmt.Errorf("postgres: empty DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}

// Repos aggregates all postgres port implementations.
type Repos struct {
	Profiles        *ProfileRepo
	Addresses       *AddressRepo
	Preferences     *PreferencesRepo
	Tags            *TagRepo
	Households      *HouseholdRepo
	Consents        *ConsentRepo
	CRM             *CRMRepo
	Segments        *SegmentRepo
	Personalization *PersonalizationRepo
	AIModels        *AIModelRepo
	Privacy         *PrivacyRepo
	Activity        *ActivityRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Profiles:        &ProfileRepo{DB: db},
		Addresses:       &AddressRepo{DB: db},
		Preferences:     &PreferencesRepo{DB: db},
		Tags:            &TagRepo{DB: db},
		Households:      &HouseholdRepo{DB: db},
		Consents:        &ConsentRepo{DB: db},
		CRM:             &CRMRepo{DB: db},
		Segments:        &SegmentRepo{DB: db},
		Personalization: &PersonalizationRepo{DB: db},
		AIModels:        &AIModelRepo{DB: db},
		Privacy:         &PrivacyRepo{DB: db},
		Activity:        &ActivityRepo{DB: db},
	}
}
