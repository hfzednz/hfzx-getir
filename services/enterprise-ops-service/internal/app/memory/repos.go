package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

type Store struct {
	mu         sync.RWMutex
	Org        map[uuid.UUID]domain.OrgNode
	OrgByCode  map[string]uuid.UUID
	Policies   map[string]domain.Policy
	Portfolios map[uuid.UUID]domain.Portfolio
	Programs   map[uuid.UUID]domain.Program
	Projects   map[uuid.UUID]domain.Project
	Milestones map[uuid.UUID]domain.Milestone
	Objectives map[uuid.UUID]domain.Objective
	KeyResults map[uuid.UUID]domain.KeyResult
	KPIs       map[uuid.UUID]domain.KPI
	Risks      map[uuid.UUID]domain.Risk
	Continuity map[string]domain.ContinuityPlan
	Audits     map[uuid.UUID]domain.AuditEngagement
	Findings   map[uuid.UUID]domain.AuditFinding
	Meetings   map[uuid.UUID]domain.Meeting
	Decisions  map[uuid.UUID]domain.Decision
	Knowledge  map[uuid.UUID]domain.KnowledgeDoc
	Resources  map[uuid.UUID]domain.ResourcePlan
	Outbox     map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Org: map[uuid.UUID]domain.OrgNode{}, OrgByCode: map[string]uuid.UUID{},
		Policies: map[string]domain.Policy{}, Portfolios: map[uuid.UUID]domain.Portfolio{},
		Programs: map[uuid.UUID]domain.Program{}, Projects: map[uuid.UUID]domain.Project{},
		Milestones: map[uuid.UUID]domain.Milestone{}, Objectives: map[uuid.UUID]domain.Objective{},
		KeyResults: map[uuid.UUID]domain.KeyResult{}, KPIs: map[uuid.UUID]domain.KPI{},
		Risks: map[uuid.UUID]domain.Risk{}, Continuity: map[string]domain.ContinuityPlan{},
		Audits: map[uuid.UUID]domain.AuditEngagement{}, Findings: map[uuid.UUID]domain.AuditFinding{},
		Meetings: map[uuid.UUID]domain.Meeting{}, Decisions: map[uuid.UUID]domain.Decision{},
		Knowledge: map[uuid.UUID]domain.KnowledgeDoc{}, Resources: map[uuid.UUID]domain.ResourcePlan{},
		Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func mk(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }

type Repos struct {
	OrgR        *OrgRepo
	PolicyR     *PolicyRepo
	PortfolioR  *PortfolioRepo
	ProgramR    *ProgramRepo
	ProjectR    *ProjectRepo
	MilestoneR  *MilestoneRepo
	ObjectiveR  *ObjectiveRepo
	KeyResultR  *KeyResultRepo
	KPIR        *KPIRepo
	RiskR       *RiskRepo
	ContinuityR *ContinuityRepo
	AuditR      *AuditRepo
	FindingR    *FindingRepo
	MeetingR    *MeetingRepo
	DecisionR   *DecisionRepo
	KnowledgeR  *KnowledgeRepo
	ResourceR   *ResourceRepo
	OutboxR     *OutboxRepo
	Security    *MockSecurity
	AI          *MockAI
	Metrics     *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		OrgR: &OrgRepo{s: s}, PolicyR: &PolicyRepo{s: s}, PortfolioR: &PortfolioRepo{s: s},
		ProgramR: &ProgramRepo{s: s}, ProjectR: &ProjectRepo{s: s}, MilestoneR: &MilestoneRepo{s: s},
		ObjectiveR: &ObjectiveRepo{s: s}, KeyResultR: &KeyResultRepo{s: s}, KPIR: &KPIRepo{s: s},
		RiskR: &RiskRepo{s: s}, ContinuityR: &ContinuityRepo{s: s}, AuditR: &AuditRepo{s: s},
		FindingR: &FindingRepo{s: s}, MeetingR: &MeetingRepo{s: s}, DecisionR: &DecisionRepo{s: s},
		KnowledgeR: &KnowledgeRepo{s: s}, ResourceR: &ResourceRepo{s: s}, OutboxR: &OutboxRepo{s: s},
		Security: &MockSecurity{allowed: true}, AI: &MockAI{}, Metrics: &MockMetrics{},
	}
}

type OrgRepo struct{ s *Store }

func (r *OrgRepo) Save(_ context.Context, n domain.OrgNode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Org[n.ID] = n
	r.s.OrgByCode[mk(n.TenantID, n.Code)] = n.ID
	return nil
}
func (r *OrgRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.OrgNode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	n, ok := r.s.Org[id]
	if !ok || n.TenantID != tenantID {
		return domain.OrgNode{}, domain.ErrNotFound
	}
	return n, nil
}
func (r *OrgRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.OrgNode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.OrgByCode[mk(tenantID, code)]
	if !ok {
		return domain.OrgNode{}, domain.ErrNotFound
	}
	return r.s.Org[id], nil
}
func (r *OrgRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.OrgNode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OrgNode{}
	for _, n := range r.s.Org {
		if n.TenantID == tenantID {
			out = append(out, n)
		}
	}
	return out, nil
}
func (r *OrgRepo) ListChildren(_ context.Context, tenantID, parentID uuid.UUID) ([]domain.OrgNode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OrgNode{}
	for _, n := range r.s.Org {
		if n.TenantID == tenantID && n.ParentID != nil && *n.ParentID == parentID {
			out = append(out, n)
		}
	}
	return out, nil
}

type PolicyRepo struct{ s *Store }

func (r *PolicyRepo) Save(_ context.Context, p domain.Policy) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Policies[mk(p.TenantID, p.Key)] = p
	return nil
}
func (r *PolicyRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Policy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Policies[mk(tenantID, key)]
	if !ok {
		return domain.Policy{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *PolicyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Policy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Policy{}
	for _, p := range r.s.Policies {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type PortfolioRepo struct{ s *Store }

func (r *PortfolioRepo) Save(_ context.Context, p domain.Portfolio) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Portfolios[p.ID] = p
	return nil
}
func (r *PortfolioRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Portfolio, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Portfolios[id]
	if !ok || p.TenantID != tenantID {
		return domain.Portfolio{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *PortfolioRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Portfolio, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Portfolio{}
	for _, p := range r.s.Portfolios {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type ProgramRepo struct{ s *Store }

func (r *ProgramRepo) Save(_ context.Context, p domain.Program) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Programs[p.ID] = p
	return nil
}
func (r *ProgramRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Program, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Programs[id]
	if !ok || p.TenantID != tenantID {
		return domain.Program{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *ProgramRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Program, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Program{}
	for _, p := range r.s.Programs {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type ProjectRepo struct{ s *Store }

func (r *ProjectRepo) Save(_ context.Context, p domain.Project) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Projects[p.ID] = p
	return nil
}
func (r *ProjectRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Project, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Projects[id]
	if !ok || p.TenantID != tenantID {
		return domain.Project{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *ProjectRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Project, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Project{}
	for _, p := range r.s.Projects {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type MilestoneRepo struct{ s *Store }

func (r *MilestoneRepo) Save(_ context.Context, m domain.Milestone) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Milestones[m.ID] = m
	return nil
}
func (r *MilestoneRepo) ListByProject(_ context.Context, tenantID, projectID uuid.UUID) ([]domain.Milestone, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Milestone{}
	for _, m := range r.s.Milestones {
		if m.TenantID == tenantID && m.ProjectID == projectID {
			out = append(out, m)
		}
	}
	return out, nil
}

type ObjectiveRepo struct{ s *Store }

func (r *ObjectiveRepo) Save(_ context.Context, o domain.Objective) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Objectives[o.ID] = o
	return nil
}
func (r *ObjectiveRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Objective, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Objective{}
	for _, o := range r.s.Objectives {
		if o.TenantID == tenantID {
			out = append(out, o)
		}
	}
	return out, nil
}

type KeyResultRepo struct{ s *Store }

func (r *KeyResultRepo) Save(_ context.Context, kr domain.KeyResult) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.KeyResults[kr.ID] = kr
	return nil
}
func (r *KeyResultRepo) ListByObjective(_ context.Context, tenantID, objectiveID uuid.UUID) ([]domain.KeyResult, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.KeyResult{}
	for _, kr := range r.s.KeyResults {
		if kr.TenantID == tenantID && kr.ObjectiveID == objectiveID {
			out = append(out, kr)
		}
	}
	return out, nil
}

type KPIRepo struct{ s *Store }

func (r *KPIRepo) Save(_ context.Context, k domain.KPI) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.KPIs[k.ID] = k
	return nil
}
func (r *KPIRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.KPI, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.KPI{}
	for _, k := range r.s.KPIs {
		if k.TenantID == tenantID {
			out = append(out, k)
		}
	}
	return out, nil
}

type RiskRepo struct{ s *Store }

func (r *RiskRepo) Save(_ context.Context, risk domain.Risk) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Risks[risk.ID] = risk
	return nil
}
func (r *RiskRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Risk, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Risk{}
	for _, risk := range r.s.Risks {
		if risk.TenantID == tenantID {
			out = append(out, risk)
		}
	}
	return out, nil
}

type ContinuityRepo struct{ s *Store }

func (r *ContinuityRepo) Save(_ context.Context, p domain.ContinuityPlan) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Continuity[mk(p.TenantID, p.Key)] = p
	return nil
}
func (r *ContinuityRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.ContinuityPlan, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Continuity[mk(tenantID, key)]
	if !ok {
		return domain.ContinuityPlan{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *ContinuityRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ContinuityPlan, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ContinuityPlan{}
	for _, p := range r.s.Continuity {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type AuditRepo struct{ s *Store }

func (r *AuditRepo) Save(_ context.Context, a domain.AuditEngagement) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Audits[a.ID] = a
	return nil
}
func (r *AuditRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.AuditEngagement, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.Audits[id]
	if !ok || a.TenantID != tenantID {
		return domain.AuditEngagement{}, domain.ErrNotFound
	}
	return a, nil
}
func (r *AuditRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.AuditEngagement, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AuditEngagement{}
	for _, a := range r.s.Audits {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type FindingRepo struct{ s *Store }

func (r *FindingRepo) Save(_ context.Context, f domain.AuditFinding) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Findings[f.ID] = f
	return nil
}
func (r *FindingRepo) ListByAudit(_ context.Context, tenantID, auditID uuid.UUID) ([]domain.AuditFinding, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AuditFinding{}
	for _, f := range r.s.Findings {
		if f.TenantID == tenantID && f.AuditID == auditID {
			out = append(out, f)
		}
	}
	return out, nil
}

type MeetingRepo struct{ s *Store }

func (r *MeetingRepo) Save(_ context.Context, m domain.Meeting) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Meetings[m.ID] = m
	return nil
}
func (r *MeetingRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Meeting, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Meeting{}
	for _, m := range r.s.Meetings {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type DecisionRepo struct{ s *Store }

func (r *DecisionRepo) Save(_ context.Context, d domain.Decision) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Decisions[d.ID] = d
	return nil
}
func (r *DecisionRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Decision, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Decision{}
	for _, d := range r.s.Decisions {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

type KnowledgeRepo struct{ s *Store }

func (r *KnowledgeRepo) Save(_ context.Context, d domain.KnowledgeDoc) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Knowledge[d.ID] = d
	return nil
}
func (r *KnowledgeRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.KnowledgeDoc, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.KnowledgeDoc{}
	for _, d := range r.s.Knowledge {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

type ResourceRepo struct{ s *Store }

func (r *ResourceRepo) Save(_ context.Context, plan domain.ResourcePlan) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Resources[plan.ID] = plan
	return nil
}
func (r *ResourceRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ResourcePlan, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ResourcePlan{}
	for _, plan := range r.s.Resources {
		if plan.TenantID == tenantID {
			out = append(out, plan)
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

type MockSecurity struct{ allowed bool }

func (m *MockSecurity) PolicyChangeAllowed(context.Context, uuid.UUID, string) (bool, error) {
	return m.allowed, nil
}

type MockAI struct{}

func (MockAI) ProjectForecast(context.Context, uuid.UUID, string) (map[string]any, error) {
	return map[string]any{"slipDays": 2, "confidence": 0.71}, nil
}
func (MockAI) RiskPrediction(context.Context, uuid.UUID) ([]string, error) {
	return []string{"capacity_shortage", "vendor_delay"}, nil
}

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
