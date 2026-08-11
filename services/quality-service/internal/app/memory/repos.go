package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/domain"
)

type Store struct {
	mu       sync.RWMutex
	Suites   map[string]domain.Suite
	Runs     map[uuid.UUID]domain.TestRun
	Results  map[uuid.UUID]domain.TestCaseResult
	Coverage map[uuid.UUID]domain.CoverageReport
	Policies map[string]domain.GatePolicy
	Evals    map[uuid.UUID]domain.GateEvaluation
	Certs    map[uuid.UUID]domain.Certification
	Flaky    map[string]domain.FlakyRecord
	Perf     map[uuid.UUID]domain.PerfMetric
	Security map[uuid.UUID]domain.SecurityFinding
	Outbox   map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Suites: map[string]domain.Suite{}, Runs: map[uuid.UUID]domain.TestRun{},
		Results: map[uuid.UUID]domain.TestCaseResult{}, Coverage: map[uuid.UUID]domain.CoverageReport{},
		Policies: map[string]domain.GatePolicy{}, Evals: map[uuid.UUID]domain.GateEvaluation{},
		Certs: map[uuid.UUID]domain.Certification{}, Flaky: map[string]domain.FlakyRecord{},
		Perf: map[uuid.UUID]domain.PerfMetric{}, Security: map[uuid.UUID]domain.SecurityFinding{},
		Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func suiteKey(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }
func flakyKey(tenantID uuid.UUID, suite, name string) string {
	return tenantID.String() + ":" + suite + ":" + name
}

type Repos struct {
	Suites   *SuiteRepo
	Runs     *RunRepo
	Results  *ResultRepo
	Coverage *CoverageRepo
	Policies *PolicyRepo
	Evals    *EvalRepo
	Certs    *CertRepo
	Flaky    *FlakyRepo
	Perf     *PerfRepo
	Security *SecurityRepo
	Outbox   *OutboxRepo
	Runner   *MockRunner
	Metrics  *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Suites: &SuiteRepo{s: s}, Runs: &RunRepo{s: s}, Results: &ResultRepo{s: s},
		Coverage: &CoverageRepo{s: s}, Policies: &PolicyRepo{s: s}, Evals: &EvalRepo{s: s},
		Certs: &CertRepo{s: s}, Flaky: &FlakyRepo{s: s}, Perf: &PerfRepo{s: s},
		Security: &SecurityRepo{s: s}, Outbox: &OutboxRepo{s: s},
		Runner: &MockRunner{}, Metrics: &MockMetrics{},
	}
}

type SuiteRepo struct{ s *Store }

func (r *SuiteRepo) Save(_ context.Context, s domain.Suite) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Suites[suiteKey(s.TenantID, s.Key)] = s
	return nil
}
func (r *SuiteRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Suite, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Suites[suiteKey(tenantID, key)]
	if !ok {
		return domain.Suite{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *SuiteRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Suite, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Suite{}
	for _, s := range r.s.Suites {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type RunRepo struct{ s *Store }

func (r *RunRepo) Save(_ context.Context, run domain.TestRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Runs[run.ID] = run
	return nil
}
func (r *RunRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.TestRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	run, ok := r.s.Runs[id]
	if !ok || run.TenantID != tenantID {
		return domain.TestRun{}, domain.ErrNotFound
	}
	return run, nil
}
func (r *RunRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.TestRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.TestRun{}
	for _, run := range r.s.Runs {
		if run.TenantID == tenantID {
			out = append(out, run)
		}
	}
	return out, nil
}

type ResultRepo struct{ s *Store }

func (r *ResultRepo) Save(_ context.Context, res domain.TestCaseResult) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Results[res.ID] = res
	return nil
}
func (r *ResultRepo) ListByRun(_ context.Context, tenantID, runID uuid.UUID) ([]domain.TestCaseResult, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.TestCaseResult{}
	for _, res := range r.s.Results {
		if res.TenantID == tenantID && res.RunID == runID {
			out = append(out, res)
		}
	}
	return out, nil
}

type CoverageRepo struct{ s *Store }

func (r *CoverageRepo) Save(_ context.Context, c domain.CoverageReport) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Coverage[c.ID] = c
	return nil
}
func (r *CoverageRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.CoverageReport, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.CoverageReport{}
	for _, c := range r.s.Coverage {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type PolicyRepo struct{ s *Store }

func (r *PolicyRepo) Save(_ context.Context, p domain.GatePolicy) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Policies[suiteKey(p.TenantID, p.Key)] = p
	return nil
}
func (r *PolicyRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.GatePolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Policies[suiteKey(tenantID, key)]
	if !ok {
		return domain.GatePolicy{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *PolicyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GatePolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GatePolicy{}
	for _, p := range r.s.Policies {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type EvalRepo struct{ s *Store }

func (r *EvalRepo) Save(_ context.Context, e domain.GateEvaluation) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Evals[e.ID] = e
	return nil
}
func (r *EvalRepo) ListByRun(_ context.Context, tenantID, runID uuid.UUID) ([]domain.GateEvaluation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GateEvaluation{}
	for _, e := range r.s.Evals {
		if e.TenantID == tenantID && e.RunID == runID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *EvalRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GateEvaluation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GateEvaluation{}
	for _, e := range r.s.Evals {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

type CertRepo struct{ s *Store }

func (r *CertRepo) Save(_ context.Context, c domain.Certification) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Certs[c.ID] = c
	return nil
}
func (r *CertRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Certification, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Certs[id]
	if !ok || c.TenantID != tenantID {
		return domain.Certification{}, domain.ErrNotFound
	}
	return c, nil
}
func (r *CertRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Certification, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Certification{}
	for _, c := range r.s.Certs {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type FlakyRepo struct{ s *Store }

func (r *FlakyRepo) Save(_ context.Context, f domain.FlakyRecord) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Flaky[flakyKey(f.TenantID, f.SuiteKey, f.TestName)] = f
	return nil
}
func (r *FlakyRepo) GetByName(_ context.Context, tenantID uuid.UUID, suiteKey, name string) (domain.FlakyRecord, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	f, ok := r.s.Flaky[flakyKey(tenantID, suiteKey, name)]
	if !ok {
		return domain.FlakyRecord{}, domain.ErrNotFound
	}
	return f, nil
}
func (r *FlakyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.FlakyRecord, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.FlakyRecord{}
	for _, f := range r.s.Flaky {
		if f.TenantID == tenantID {
			out = append(out, f)
		}
	}
	return out, nil
}

type PerfRepo struct{ s *Store }

func (r *PerfRepo) Save(_ context.Context, p domain.PerfMetric) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Perf[p.ID] = p
	return nil
}
func (r *PerfRepo) ListByRun(_ context.Context, tenantID, runID uuid.UUID) ([]domain.PerfMetric, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.PerfMetric{}
	for _, p := range r.s.Perf {
		if p.TenantID == tenantID && p.RunID == runID {
			out = append(out, p)
		}
	}
	return out, nil
}

type SecurityRepo struct{ s *Store }

func (r *SecurityRepo) Save(_ context.Context, f domain.SecurityFinding) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Security[f.ID] = f
	return nil
}
func (r *SecurityRepo) ListByRun(_ context.Context, tenantID, runID uuid.UUID) ([]domain.SecurityFinding, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SecurityFinding{}
	for _, f := range r.s.Security {
		if f.TenantID == tenantID && f.RunID == runID {
			out = append(out, f)
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}
func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}

type MockRunner struct{}

func (MockRunner) Dispatch(context.Context, domain.Suite, domain.TestRun) error { return nil }

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
