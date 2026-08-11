package postgres

import "database/sql"

// Repos groups liveops persistence adapters.
type Repos struct {
	Flags       *FlagRepo
	Configs     *ConfigRepo
	Experiments *ExperimentRepo
	Events      *EventRepo
	Changes     *ChangeRepo
	Rollbacks   *RollbackRepo
	Outbox      *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Flags:       &FlagRepo{DB: db},
		Configs:     &ConfigRepo{DB: db},
		Experiments: &ExperimentRepo{DB: db},
		Events:      &EventRepo{DB: db},
		Changes:     &ChangeRepo{DB: db},
		Rollbacks:   &RollbackRepo{DB: db},
		Outbox:      &OutboxRepo{DB: db},
	}
}
