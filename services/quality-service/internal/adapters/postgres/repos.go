package postgres

import "database/sql"

type Repos struct {
	Suites    *SuiteRepo
	Runs      *RunRepo
	Results   *ResultRepo
	Coverage  *CoverageRepo
	Policies  *PolicyRepo
	Evals     *EvalRepo
	Certs     *CertRepo
	Flaky     *FlakyRepo
	Perf      *PerfRepo
	Security  *SecurityRepo
	Outbox    *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Suites: &SuiteRepo{DB: db}, Runs: &RunRepo{DB: db}, Results: &ResultRepo{DB: db},
		Coverage: &CoverageRepo{DB: db}, Policies: &PolicyRepo{DB: db}, Evals: &EvalRepo{DB: db},
		Certs: &CertRepo{DB: db}, Flaky: &FlakyRepo{DB: db}, Perf: &PerfRepo{DB: db},
		Security: &SecurityRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
