package postgres

import "database/sql"

// Repos groups dispatch persistence adapters.
type Repos struct {
	Dispatches *DispatchRepo
	Couriers   *CourierPool
	Vehicles   *VehicleRepo
	Outbox     *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Dispatches: &DispatchRepo{DB: db},
		Couriers:   &CourierPool{DB: db},
		Vehicles:   &VehicleRepo{DB: db},
		Outbox:     &OutboxRepo{DB: db},
	}
}
