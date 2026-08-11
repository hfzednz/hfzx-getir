package postgres

import "database/sql"

// Repos groups geofence persistence adapters.
type Repos struct {
	Zones  *ZoneRepo
	Outbox *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Zones:  &ZoneRepo{DB: db},
		Outbox: &OutboxRepo{DB: db},
	}
}
