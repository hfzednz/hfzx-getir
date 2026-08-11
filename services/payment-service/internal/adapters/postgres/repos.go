package postgres

import "database/sql"

// Repos aggregates postgres port implementations.
type Repos struct {
	Intents *IntentRepo
	Outbox  *OutboxRepo
}

// NewRepos wires intent and outbox repositories on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Intents: &IntentRepo{DB: db},
		Outbox:  &OutboxRepo{DB: db},
	}
}
