package postgres

import "database/sql"

// Repos groups tracking persistence adapters.
type Repos struct {
	Locations *LocationRepo
	Timelines *TimelineRepo
	Outbox    *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Locations: &LocationRepo{DB: db},
		Timelines: &TimelineRepo{DB: db},
		Outbox:    &OutboxRepo{DB: db},
	}
}
