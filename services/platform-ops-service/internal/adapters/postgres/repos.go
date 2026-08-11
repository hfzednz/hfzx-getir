package postgres

import "database/sql"

// Repos groups platform-ops persistence adapters.
type Repos struct {
	Deployments *DeploymentRepo
	Scaling     *ScalingRepo
	Backups     *BackupRepo
	Recoveries  *RecoveryRepo
	Alerts      *AlertRepo
	Costs       *CostRepo
	SLOs        *SLORepo
	Outbox      *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Deployments: &DeploymentRepo{DB: db},
		Scaling:     &ScalingRepo{DB: db},
		Backups:     &BackupRepo{DB: db},
		Recoveries:  &RecoveryRepo{DB: db},
		Alerts:      &AlertRepo{DB: db},
		Costs:       &CostRepo{DB: db},
		SLOs:        &SLORepo{DB: db},
		Outbox:      &OutboxRepo{DB: db},
	}
}
