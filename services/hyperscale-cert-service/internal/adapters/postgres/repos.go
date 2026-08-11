package postgres

import "database/sql"

type Repos struct {
	Audits       *AuditRepo
	Findings     *FindingRepo
	Benchmarks   *BenchmarkRepo
	Capacity     *CapacityRepo
	Chaos        *ChaosRepo
	Tuning       *TuningRepo
	Certificates *CertificateRepo
	Outbox       *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Audits: &AuditRepo{DB: db}, Findings: &FindingRepo{DB: db}, Benchmarks: &BenchmarkRepo{DB: db},
		Capacity: &CapacityRepo{DB: db}, Chaos: &ChaosRepo{DB: db}, Tuning: &TuningRepo{DB: db},
		Certificates: &CertificateRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
