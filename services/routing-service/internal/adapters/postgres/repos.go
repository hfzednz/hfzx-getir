package postgres

import "database/sql"

// Repos groups routing persistence adapters.
type Repos struct {
	Routes *RouteRepo
	Outbox *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Routes: &RouteRepo{DB: db},
		Outbox: &OutboxRepo{DB: db},
	}
}
