// Package postgres implements production Catalog Service repositories against PostgreSQL.
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
	Products   *ProductRepo
	Variants   *VariantRepo
	SKUs       *SKURepo
	Categories *CategoryRepo
	Brands     *BrandRepo
	Attributes *AttributeRepo
	Locales    *LocaleRepo
	SEO        *SEORepo
	Media      *MediaRepo
	Bundles    *BundleRepo
	Relations  *RelationRepo
	Versions   *VersionRepo
	Workflow   *WorkflowRepo
	ImportJobs *ImportJobRepo
	Compliance *ComplianceRepo
	Suppliers  *SupplierRepo
}

// NewRepos wires all repository ports on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Products:   &ProductRepo{DB: db},
		Variants:   &VariantRepo{DB: db},
		SKUs:       &SKURepo{DB: db},
		Categories: &CategoryRepo{DB: db},
		Brands:     &BrandRepo{DB: db},
		Attributes: &AttributeRepo{DB: db},
		Locales:    &LocaleRepo{DB: db},
		SEO:        &SEORepo{DB: db},
		Media:      &MediaRepo{DB: db},
		Bundles:    &BundleRepo{DB: db},
		Relations:  &RelationRepo{DB: db},
		Versions:   &VersionRepo{DB: db},
		Workflow:   &WorkflowRepo{DB: db},
		ImportJobs: &ImportJobRepo{DB: db},
		Compliance: &ComplianceRepo{DB: db},
		Suppliers:  &SupplierRepo{DB: db},
	}
}
