package postgres

import "database/sql"

// Repos groups pricing persistence adapters.
type Repos struct {
	Prices   *PriceRepo
	Taxes    *TaxRepo
	Dynamics *DynamicRepo
	Audits   *QuoteAuditRepo
	Outbox   *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Prices:   &PriceRepo{DB: db},
		Taxes:    &TaxRepo{DB: db},
		Dynamics: &DynamicRepo{DB: db},
		Audits:   &QuoteAuditRepo{DB: db},
		Outbox:   &OutboxRepo{DB: db},
	}
}
