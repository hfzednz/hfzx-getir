package postgres

import "database/sql"

// Repos groups data-platform persistence adapters.
type Repos struct {
	Schemas     *SchemaRepo
	Events      *EventRepo
	Streams     *StreamRepo
	Lake        *LakeRepo
	Warehouse   *WarehouseRepo
	Realtime    *RealtimeRepo
	Experiments *ExperimentRepo
	Reports     *ReportRepo
	Obs         *ObsRepo
	Alerts      *AlertRepo
	Catalog     *CatalogRepo
	Quality     *QualityRepo
	Outbox      *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Schemas: &SchemaRepo{DB: db}, Events: &EventRepo{DB: db}, Streams: &StreamRepo{DB: db},
		Lake: &LakeRepo{DB: db}, Warehouse: &WarehouseRepo{DB: db}, Realtime: &RealtimeRepo{DB: db},
		Experiments: &ExperimentRepo{DB: db}, Reports: &ReportRepo{DB: db}, Obs: &ObsRepo{DB: db},
		Alerts: &AlertRepo{DB: db}, Catalog: &CatalogRepo{DB: db}, Quality: &QualityRepo{DB: db},
		Outbox: &OutboxRepo{DB: db},
	}
}
