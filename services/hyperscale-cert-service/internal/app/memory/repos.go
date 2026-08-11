package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
)

type Store struct {
	mu             sync.RWMutex
	Audits         map[uuid.UUID]domain.Audit
	Findings       map[uuid.UUID]domain.Finding
	Benchmarks     map[uuid.UUID]domain.BenchmarkRun
	BenchmarkOrder []uuid.UUID // insertion order for LatestByKind (clock-resolution safe)
	Capacity       map[uuid.UUID]domain.CapacityScenario
	Chaos          map[uuid.UUID]domain.ChaosExperiment
	Tuning         map[string]domain.TuningProfile
	Certificates   map[uuid.UUID]domain.Certificate
	Outbox         map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Audits: map[uuid.UUID]domain.Audit{}, Findings: map[uuid.UUID]domain.Finding{},
		Benchmarks: map[uuid.UUID]domain.BenchmarkRun{}, Capacity: map[uuid.UUID]domain.CapacityScenario{},
		Chaos: map[uuid.UUID]domain.ChaosExperiment{}, Tuning: map[string]domain.TuningProfile{},
		Certificates: map[uuid.UUID]domain.Certificate{}, Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func mk(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }

type Repos struct {
	AuditR   *AuditRepo
	FindingR *FindingRepo
	BenchR   *BenchmarkRepo
	CapR     *CapacityRepo
	ChaosR   *ChaosRepo
	TuningR  *TuningRepo
	CertR    *CertificateRepo
	OutboxR  *OutboxRepo
	Quality  *MockQuality
	Platform *MockPlatform
	Security *MockSecurity
	Metrics  *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		AuditR: &AuditRepo{s: s}, FindingR: &FindingRepo{s: s}, BenchR: &BenchmarkRepo{s: s},
		CapR: &CapacityRepo{s: s}, ChaosR: &ChaosRepo{s: s}, TuningR: &TuningRepo{s: s},
		CertR: &CertificateRepo{s: s}, OutboxR: &OutboxRepo{s: s},
		Quality: &MockQuality{ok: true}, Platform: &MockPlatform{ok: true},
		Security: &MockSecurity{ok: true}, Metrics: &MockMetrics{},
	}
}

type AuditRepo struct{ s *Store }

func (r *AuditRepo) Save(_ context.Context, a domain.Audit) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Audits[a.ID] = a
	return nil
}
func (r *AuditRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Audit, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.Audits[id]
	if !ok || a.TenantID != tenantID {
		return domain.Audit{}, domain.ErrNotFound
	}
	return a, nil
}
func (r *AuditRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Audit, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Audit{}
	for _, a := range r.s.Audits {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type FindingRepo struct{ s *Store }

func (r *FindingRepo) Save(_ context.Context, f domain.Finding) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Findings[f.ID] = f
	return nil
}
func (r *FindingRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Finding, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Finding{}
	for _, f := range r.s.Findings {
		if f.TenantID == tenantID {
			out = append(out, f)
		}
	}
	return out, nil
}
func (r *FindingRepo) ListByAudit(_ context.Context, tenantID, auditID uuid.UUID) ([]domain.Finding, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Finding{}
	for _, f := range r.s.Findings {
		if f.TenantID == tenantID && f.AuditID == auditID {
			out = append(out, f)
		}
	}
	return out, nil
}
func (r *FindingRepo) OpenCritical(_ context.Context, tenantID uuid.UUID) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, f := range r.s.Findings {
		if f.TenantID == tenantID && f.Status == "open" && f.Severity == domain.SeverityCritical {
			n++
		}
	}
	return n, nil
}

type BenchmarkRepo struct{ s *Store }

func (r *BenchmarkRepo) Save(_ context.Context, b domain.BenchmarkRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Benchmarks[b.ID] = b
	r.s.BenchmarkOrder = append(r.s.BenchmarkOrder, b.ID)
	return nil
}
func (r *BenchmarkRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.BenchmarkRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.BenchmarkRun{}
	for _, b := range r.s.Benchmarks {
		if b.TenantID == tenantID {
			out = append(out, b)
		}
	}
	return out, nil
}
func (r *BenchmarkRepo) LatestByKind(_ context.Context, tenantID uuid.UUID, kind domain.BenchmarkKind) (domain.BenchmarkRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	// Walk newest-first so equal CreatedAt (coarse OS clocks) still picks last write.
	for i := len(r.s.BenchmarkOrder) - 1; i >= 0; i-- {
		b, ok := r.s.Benchmarks[r.s.BenchmarkOrder[i]]
		if !ok {
			continue
		}
		if b.TenantID == tenantID && b.Kind == kind {
			return b, nil
		}
	}
	return domain.BenchmarkRun{}, domain.ErrNotFound
}

type CapacityRepo struct{ s *Store }

func (r *CapacityRepo) Save(_ context.Context, c domain.CapacityScenario) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Capacity[c.ID] = c
	return nil
}
func (r *CapacityRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.CapacityScenario, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.CapacityScenario{}
	for _, c := range r.s.Capacity {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type ChaosRepo struct{ s *Store }

func (r *ChaosRepo) Save(_ context.Context, c domain.ChaosExperiment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Chaos[c.ID] = c
	return nil
}
func (r *ChaosRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ChaosExperiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Chaos[id]
	if !ok || c.TenantID != tenantID {
		return domain.ChaosExperiment{}, domain.ErrNotFound
	}
	return c, nil
}
func (r *ChaosRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ChaosExperiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ChaosExperiment{}
	for _, c := range r.s.Chaos {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type TuningRepo struct{ s *Store }

func (r *TuningRepo) Save(_ context.Context, t domain.TuningProfile) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Tuning[mk(t.TenantID, t.Key)] = t
	return nil
}
func (r *TuningRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.TuningProfile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	t, ok := r.s.Tuning[mk(tenantID, key)]
	if !ok {
		return domain.TuningProfile{}, domain.ErrNotFound
	}
	return t, nil
}
func (r *TuningRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.TuningProfile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.TuningProfile{}
	for _, t := range r.s.Tuning {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

type CertificateRepo struct{ s *Store }

func (r *CertificateRepo) Save(_ context.Context, c domain.Certificate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Certificates[c.ID] = c
	return nil
}
func (r *CertificateRepo) Latest(_ context.Context, tenantID uuid.UUID) (domain.Certificate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var best domain.Certificate
	found := false
	for _, c := range r.s.Certificates {
		if c.TenantID == tenantID {
			if !found || c.CreatedAt.After(best.CreatedAt) {
				best = c
				found = true
			}
		}
	}
	if !found {
		return domain.Certificate{}, domain.ErrNotFound
	}
	return best, nil
}
func (r *CertificateRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Certificate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Certificate{}
	for _, c := range r.s.Certificates {
		if c.TenantID == tenantID {
			out = append(out, c)
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

type MockQuality struct{ ok bool }

func (m *MockQuality) ReleaseGatesGreen(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockPlatform struct{ ok bool }

func (m *MockPlatform) DRDrillPassed(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockSecurity struct{ ok bool }

func (m *MockSecurity) ZeroCriticalVulns(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
