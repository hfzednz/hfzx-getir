package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Audits       map[uuid.UUID]domain.AutonomyAudit
	Weaknesses   map[uuid.UUID]domain.Weakness
	Heals        map[uuid.UUID]domain.HealAction
	Reviews      map[uuid.UUID]domain.AICTOReview
	Evolution    map[uuid.UUID]domain.EvolutionTask
	Releases     map[uuid.UUID]domain.ReleasePlan
	Governance   map[uuid.UUID]domain.GovernanceLoop
	Assistants   map[uuid.UUID]domain.ExecutiveAssistant
	Teams        map[uuid.UUID]domain.DigitalTeam
	Dependencies map[uuid.UUID]domain.DependencyEdge
	Genesis      map[uuid.UUID]domain.GenesisCertificate
	Outbox       map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Audits: map[uuid.UUID]domain.AutonomyAudit{}, Weaknesses: map[uuid.UUID]domain.Weakness{},
		Heals: map[uuid.UUID]domain.HealAction{}, Reviews: map[uuid.UUID]domain.AICTOReview{},
		Evolution: map[uuid.UUID]domain.EvolutionTask{}, Releases: map[uuid.UUID]domain.ReleasePlan{},
		Governance: map[uuid.UUID]domain.GovernanceLoop{}, Assistants: map[uuid.UUID]domain.ExecutiveAssistant{},
		Teams: map[uuid.UUID]domain.DigitalTeam{}, Dependencies: map[uuid.UUID]domain.DependencyEdge{},
		Genesis: map[uuid.UUID]domain.GenesisCertificate{}, Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

type Repos struct {
	Audits       *AuditRepo
	Weaknesses   *WeaknessRepo
	Heals        *HealRepo
	Reviews      *ReviewRepo
	Evolution    *EvolutionRepo
	Releases     *ReleaseRepo
	Governance   *GovernanceRepo
	Assistants   *AssistantRepo
	Teams        *TeamRepo
	Dependencies *DependencyRepo
	Genesis      *GenesisRepo
	Outbox       *OutboxRepo
	Hyperscale   *MockHyperscale
	Platform     *MockPlatform
	Quality      *MockQuality
	Security     *MockSecurity
	LiveOps      *MockLiveOps
	Metrics      *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Audits: &AuditRepo{s: s}, Weaknesses: &WeaknessRepo{s: s}, Heals: &HealRepo{s: s},
		Reviews: &ReviewRepo{s: s}, Evolution: &EvolutionRepo{s: s}, Releases: &ReleaseRepo{s: s},
		Governance: &GovernanceRepo{s: s}, Assistants: &AssistantRepo{s: s}, Teams: &TeamRepo{s: s},
		Dependencies: &DependencyRepo{s: s}, Genesis: &GenesisRepo{s: s}, Outbox: &OutboxRepo{s: s},
		Hyperscale: &MockHyperscale{ok: true}, Platform: &MockPlatform{}, Quality: &MockQuality{ok: true},
		Security: &MockSecurity{ok: true}, LiveOps: &MockLiveOps{ok: true}, Metrics: &MockMetrics{},
	}
}

type AuditRepo struct{ s *Store }

func (r *AuditRepo) Save(_ context.Context, a domain.AutonomyAudit) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Audits[a.ID] = a
	return nil
}
func (r *AuditRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.AutonomyAudit, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AutonomyAudit{}
	for _, a := range r.s.Audits {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type WeaknessRepo struct{ s *Store }

func (r *WeaknessRepo) Save(_ context.Context, w domain.Weakness) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Weaknesses[w.ID] = w
	return nil
}
func (r *WeaknessRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Weakness, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Weakness{}
	for _, w := range r.s.Weaknesses {
		if w.TenantID == tenantID {
			out = append(out, w)
		}
	}
	return out, nil
}
func (r *WeaknessRepo) OpenCount(_ context.Context, tenantID uuid.UUID) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, w := range r.s.Weaknesses {
		if w.TenantID == tenantID && w.Status == "open" {
			n++
		}
	}
	return n, nil
}

type HealRepo struct{ s *Store }

func (r *HealRepo) Save(_ context.Context, a domain.HealAction) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Heals[a.ID] = a
	return nil
}
func (r *HealRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.HealAction, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.HealAction{}
	for _, a := range r.s.Heals {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *HealRepo) ExecutedCount(_ context.Context, tenantID uuid.UUID) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n := 0
	for _, a := range r.s.Heals {
		if a.TenantID == tenantID && a.Status == "executed" {
			n++
		}
	}
	return n, nil
}

type ReviewRepo struct{ s *Store }

func (r *ReviewRepo) Save(_ context.Context, rev domain.AICTOReview) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Reviews[rev.ID] = rev
	return nil
}
func (r *ReviewRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.AICTOReview, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AICTOReview{}
	for _, rev := range r.s.Reviews {
		if rev.TenantID == tenantID {
			out = append(out, rev)
		}
	}
	return out, nil
}

type EvolutionRepo struct{ s *Store }

func (r *EvolutionRepo) Save(_ context.Context, t domain.EvolutionTask) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Evolution[t.ID] = t
	return nil
}
func (r *EvolutionRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.EvolutionTask, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.EvolutionTask{}
	for _, t := range r.s.Evolution {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

type ReleaseRepo struct{ s *Store }

func (r *ReleaseRepo) Save(_ context.Context, p domain.ReleasePlan) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Releases[p.ID] = p
	return nil
}
func (r *ReleaseRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ReleasePlan, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ReleasePlan{}
	for _, p := range r.s.Releases {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type GovernanceRepo struct{ s *Store }

func (r *GovernanceRepo) Save(_ context.Context, g domain.GovernanceLoop) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Governance[g.ID] = g
	return nil
}
func (r *GovernanceRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GovernanceLoop, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GovernanceLoop{}
	for _, g := range r.s.Governance {
		if g.TenantID == tenantID {
			out = append(out, g)
		}
	}
	return out, nil
}

type AssistantRepo struct{ s *Store }

func (r *AssistantRepo) Save(_ context.Context, a domain.ExecutiveAssistant) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Assistants[a.ID] = a
	return nil
}
func (r *AssistantRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ExecutiveAssistant, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ExecutiveAssistant{}
	for _, a := range r.s.Assistants {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type TeamRepo struct{ s *Store }

func (r *TeamRepo) Save(_ context.Context, t domain.DigitalTeam) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Teams[t.ID] = t
	return nil
}
func (r *TeamRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DigitalTeam, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DigitalTeam{}
	for _, t := range r.s.Teams {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

type DependencyRepo struct{ s *Store }

func (r *DependencyRepo) Save(_ context.Context, e domain.DependencyEdge) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Dependencies[e.ID] = e
	return nil
}
func (r *DependencyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DependencyEdge, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DependencyEdge{}
	for _, e := range r.s.Dependencies {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

type GenesisRepo struct{ s *Store }

func (r *GenesisRepo) Save(_ context.Context, c domain.GenesisCertificate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Genesis[c.ID] = c
	return nil
}
func (r *GenesisRepo) Latest(_ context.Context, tenantID uuid.UUID) (domain.GenesisCertificate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var best domain.GenesisCertificate
	found := false
	for _, c := range r.s.Genesis {
		if c.TenantID == tenantID {
			if !found || c.CreatedAt.After(best.CreatedAt) {
				best = c
				found = true
			}
		}
	}
	if !found {
		return domain.GenesisCertificate{}, domain.ErrNotFound
	}
	return best, nil
}
func (r *GenesisRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GenesisCertificate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GenesisCertificate{}
	for _, c := range r.s.Genesis {
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

type MockHyperscale struct{ ok bool }

func (m *MockHyperscale) Certified(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }
func (m *MockHyperscale) Set(ok bool)                                      { m.ok = ok }

type MockPlatform struct{}

func (MockPlatform) ExecuteHeal(context.Context, uuid.UUID, string, string) error { return nil }

type MockQuality struct{ ok bool }

func (m *MockQuality) Healthy(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockSecurity struct{ ok bool }

func (m *MockSecurity) Healthy(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockLiveOps struct{ ok bool }

func (m *MockLiveOps) EmergencyOffClear(context.Context, uuid.UUID) (bool, error) { return m.ok, nil }

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
