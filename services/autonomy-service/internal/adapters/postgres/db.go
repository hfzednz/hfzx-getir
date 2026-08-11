package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("postgres: empty DATABASE_URL")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
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
	Audits       *AuditRepo
	Weaknesses   *WeaknessRepo
	Heals        *HealRepo
	Reviews      *ReviewRepo
	Evolution    *EvolutionRepo
	Releases     *ReleaseRepo
	Governance   *GovernanceRepo
	Assistants   *AssistantRepo
	Teams        *TeamRepo
	Dependencies *DependencyRepo
	Genesis      *GenesisRepo
	Outbox       *OutboxRepo
}

// NewRepos wires all repository ports on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Audits:       &AuditRepo{DB: db},
		Weaknesses:   &WeaknessRepo{DB: db},
		Heals:        &HealRepo{DB: db},
		Reviews:      &ReviewRepo{DB: db},
		Evolution:    &EvolutionRepo{DB: db},
		Releases:     &ReleaseRepo{DB: db},
		Governance:   &GovernanceRepo{DB: db},
		Assistants:   &AssistantRepo{DB: db},
		Teams:        &TeamRepo{DB: db},
		Dependencies: &DependencyRepo{DB: db},
		Genesis:      &GenesisRepo{DB: db},
		Outbox:       &OutboxRepo{DB: db},
	}
}
