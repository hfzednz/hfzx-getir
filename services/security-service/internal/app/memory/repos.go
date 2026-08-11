package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	Policies  []domain.SecurityPolicy
	Decisions []domain.PolicyDecision
	Audits    []domain.AuditEvent
	Secrets   []domain.SecretMeta
	Threats   []domain.ThreatAlert
	Vulns     []domain.ScanFinding
	Incidents []domain.Incident
	Controls  []domain.ComplianceControl
	Evidence  []domain.ComplianceEvidence
	AuditsRun []domain.ComplianceAuditRun
	Assets    []domain.DataAsset
	Privacy   []domain.PrivacyRequest
	Risks     []domain.RiskItem
	Access    []domain.AccessRequest
	Devices   []domain.DeviceTrust
	AISec     []domain.AISecurityEvent
	Fraud     []domain.FraudSignal
	Outbox    []domain.OutboxMessage
}

func NewStore() *Store { return &Store{} }

type Repos struct {
	Policies   *PolicyRepo
	Audits     *AuditRepo
	Secrets    *SecretRepo
	Threats    *ThreatRepo
	Vulns      *VulnRepo
	Incidents  *IncidentRepo
	Compliance *ComplianceRepo
	DataGov    *DataGovRepo
	Risks      *RiskRepo
	Access     *AccessRepo
	Devices    *DeviceRepo
	AISec      *AISecRepo
	FraudSigs  *FraudRepo
	Outbox     *OutboxRepo
	Vault      *MockVault
	OPA        *MockOPA
	Identity   *MockIdentity
	Fraud      *MockFraud
	SIEM       *MockSIEM
	SOAR       *MockSOAR
	AIGuard    *MockAIGuard
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Policies: &PolicyRepo{s: s}, Audits: &AuditRepo{s: s}, Secrets: &SecretRepo{s: s},
		Threats: &ThreatRepo{s: s}, Vulns: &VulnRepo{s: s}, Incidents: &IncidentRepo{s: s},
		Compliance: &ComplianceRepo{s: s}, DataGov: &DataGovRepo{s: s}, Risks: &RiskRepo{s: s},
		Access: &AccessRepo{s: s}, Devices: &DeviceRepo{s: s}, AISec: &AISecRepo{s: s},
		FraudSigs: &FraudRepo{s: s}, Outbox: &OutboxRepo{s: s},
		Vault: &MockVault{Version: 1}, OPA: &MockOPA{}, Identity: &MockIdentity{},
		Fraud: &MockFraud{}, SIEM: &MockSIEM{}, SOAR: &MockSOAR{}, AIGuard: &MockAIGuard{},
	}
}

type PolicyRepo struct{ s *Store }

func (r *PolicyRepo) Save(_ context.Context, p domain.SecurityPolicy) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Policies {
		if r.s.Policies[i].ID == p.ID {
			r.s.Policies[i] = p
			return nil
		}
	}
	r.s.Policies = append(r.s.Policies, p)
	return nil
}

func (r *PolicyRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.SecurityPolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	key = domain.NormalizeKey(key)
	for _, p := range r.s.Policies {
		if p.TenantID == tenantID && p.Key == key {
			return p, nil
		}
	}
	return domain.SecurityPolicy{}, domain.ErrNotFound
}

func (r *PolicyRepo) List(_ context.Context, tenantID uuid.UUID, kind string) ([]domain.SecurityPolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SecurityPolicy{}
	for _, p := range r.s.Policies {
		if p.TenantID == tenantID && (kind == "" || p.Kind == kind) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *PolicyRepo) SaveDecision(_ context.Context, d domain.PolicyDecision) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Decisions = append(r.s.Decisions, d)
	return nil
}

type AuditRepo struct{ s *Store }

func (r *AuditRepo) Append(_ context.Context, e domain.AuditEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Audits = append(r.s.Audits, e)
	return nil
}

func (r *AuditRepo) LastHash(_ context.Context, tenantID uuid.UUID) (string, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for i := len(r.s.Audits) - 1; i >= 0; i-- {
		if r.s.Audits[i].TenantID == tenantID {
			return r.s.Audits[i].Hash, nil
		}
	}
	return "", nil
}

func (r *AuditRepo) Search(_ context.Context, tenantID uuid.UUID, action, actor string, limit int) ([]domain.AuditEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := []domain.AuditEvent{}
	for i := len(r.s.Audits) - 1; i >= 0 && len(out) < limit; i-- {
		e := r.s.Audits[i]
		if e.TenantID != tenantID {
			continue
		}
		if action != "" && e.Action != action {
			continue
		}
		if actor != "" && e.ActorID != actor {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

type SecretRepo struct{ s *Store }

func (r *SecretRepo) Save(_ context.Context, s domain.SecretMeta) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Secrets {
		if r.s.Secrets[i].ID == s.ID {
			r.s.Secrets[i] = s
			return nil
		}
	}
	r.s.Secrets = append(r.s.Secrets, s)
	return nil
}

func (r *SecretRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.SecretMeta, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, s := range r.s.Secrets {
		if s.TenantID == tenantID && s.ID == id {
			return s, nil
		}
	}
	return domain.SecretMeta{}, domain.ErrNotFound
}

func (r *SecretRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.SecretMeta, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SecretMeta{}
	for _, s := range r.s.Secrets {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type ThreatRepo struct{ s *Store }

func (r *ThreatRepo) Save(_ context.Context, t domain.ThreatAlert) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Threats {
		if r.s.Threats[i].ID == t.ID {
			r.s.Threats[i] = t
			return nil
		}
	}
	r.s.Threats = append(r.s.Threats, t)
	return nil
}

func (r *ThreatRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ThreatAlert, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, t := range r.s.Threats {
		if t.TenantID == tenantID && t.ID == id {
			return t, nil
		}
	}
	return domain.ThreatAlert{}, domain.ErrNotFound
}

func (r *ThreatRepo) List(_ context.Context, tenantID uuid.UUID, status string) ([]domain.ThreatAlert, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ThreatAlert{}
	for _, t := range r.s.Threats {
		if t.TenantID == tenantID && (status == "" || t.Status == status) {
			out = append(out, t)
		}
	}
	return out, nil
}

type VulnRepo struct{ s *Store }

func (r *VulnRepo) Save(_ context.Context, f domain.ScanFinding) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Vulns {
		if r.s.Vulns[i].ID == f.ID {
			r.s.Vulns[i] = f
			return nil
		}
	}
	r.s.Vulns = append(r.s.Vulns, f)
	return nil
}

func (r *VulnRepo) List(_ context.Context, tenantID uuid.UUID, status string) ([]domain.ScanFinding, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ScanFinding{}
	for _, f := range r.s.Vulns {
		if f.TenantID == tenantID && (status == "" || f.Status == status) {
			out = append(out, f)
		}
	}
	return out, nil
}

type IncidentRepo struct{ s *Store }

func (r *IncidentRepo) Save(_ context.Context, i domain.Incident) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for j := range r.s.Incidents {
		if r.s.Incidents[j].ID == i.ID {
			r.s.Incidents[j] = i
			return nil
		}
	}
	r.s.Incidents = append(r.s.Incidents, i)
	return nil
}

func (r *IncidentRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Incident, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, i := range r.s.Incidents {
		if i.TenantID == tenantID && i.ID == id {
			return i, nil
		}
	}
	return domain.Incident{}, domain.ErrNotFound
}

func (r *IncidentRepo) List(_ context.Context, tenantID uuid.UUID, status string) ([]domain.Incident, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Incident{}
	for _, i := range r.s.Incidents {
		if i.TenantID == tenantID && (status == "" || i.Status == status) {
			out = append(out, i)
		}
	}
	return out, nil
}

type ComplianceRepo struct{ s *Store }

func (r *ComplianceRepo) SaveControl(_ context.Context, c domain.ComplianceControl) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Controls {
		if r.s.Controls[i].ID == c.ID {
			r.s.Controls[i] = c
			return nil
		}
	}
	r.s.Controls = append(r.s.Controls, c)
	return nil
}

func (r *ComplianceRepo) ListControls(_ context.Context, tenantID uuid.UUID, framework string) ([]domain.ComplianceControl, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	framework = domain.NormalizeKey(framework)
	out := []domain.ComplianceControl{}
	for _, c := range r.s.Controls {
		if c.TenantID == tenantID && (framework == "" || c.Framework == framework) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *ComplianceRepo) SaveEvidence(_ context.Context, e domain.ComplianceEvidence) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Evidence = append(r.s.Evidence, e)
	for i := range r.s.Controls {
		if r.s.Controls[i].ID == e.ControlID {
			r.s.Controls[i].EvidenceIDs = append(r.s.Controls[i].EvidenceIDs, e.ID)
			if r.s.Controls[i].Status == "not_started" {
				r.s.Controls[i].Status = "in_progress"
			}
		}
	}
	return nil
}

func (r *ComplianceRepo) SaveAuditRun(_ context.Context, run domain.ComplianceAuditRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AuditsRun = append(r.s.AuditsRun, run)
	return nil
}

func (r *ComplianceRepo) ListAuditRuns(_ context.Context, tenantID uuid.UUID) ([]domain.ComplianceAuditRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ComplianceAuditRun{}
	for _, a := range r.s.AuditsRun {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type DataGovRepo struct{ s *Store }

func (r *DataGovRepo) SaveAsset(_ context.Context, a domain.DataAsset) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Assets {
		if r.s.Assets[i].ID == a.ID {
			r.s.Assets[i] = a
			return nil
		}
	}
	r.s.Assets = append(r.s.Assets, a)
	return nil
}

func (r *DataGovRepo) ListAssets(_ context.Context, tenantID uuid.UUID) ([]domain.DataAsset, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DataAsset{}
	for _, a := range r.s.Assets {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *DataGovRepo) SavePrivacy(_ context.Context, p domain.PrivacyRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Privacy {
		if r.s.Privacy[i].ID == p.ID {
			r.s.Privacy[i] = p
			return nil
		}
	}
	r.s.Privacy = append(r.s.Privacy, p)
	return nil
}

func (r *DataGovRepo) GetPrivacy(_ context.Context, tenantID, id uuid.UUID) (domain.PrivacyRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Privacy {
		if p.TenantID == tenantID && p.ID == id {
			return p, nil
		}
	}
	return domain.PrivacyRequest{}, domain.ErrNotFound
}

func (r *DataGovRepo) ListPrivacy(_ context.Context, tenantID uuid.UUID) ([]domain.PrivacyRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.PrivacyRequest{}
	for _, p := range r.s.Privacy {
		if p.TenantID == tenantID {
			out = append(out, p)
		}
	}
	return out, nil
}

type RiskRepo struct{ s *Store }

func (r *RiskRepo) Save(_ context.Context, item domain.RiskItem) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Risks {
		if r.s.Risks[i].ID == item.ID {
			r.s.Risks[i] = item
			return nil
		}
	}
	r.s.Risks = append(r.s.Risks, item)
	return nil
}

func (r *RiskRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.RiskItem, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, item := range r.s.Risks {
		if item.TenantID == tenantID && item.ID == id {
			return item, nil
		}
	}
	return domain.RiskItem{}, domain.ErrNotFound
}

func (r *RiskRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RiskItem, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.RiskItem{}
	for _, item := range r.s.Risks {
		if item.TenantID == tenantID {
			out = append(out, item)
		}
	}
	return out, nil
}

type AccessRepo struct{ s *Store }

func (r *AccessRepo) Save(_ context.Context, a domain.AccessRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Access {
		if r.s.Access[i].ID == a.ID {
			r.s.Access[i] = a
			return nil
		}
	}
	r.s.Access = append(r.s.Access, a)
	return nil
}

func (r *AccessRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.AccessRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, a := range r.s.Access {
		if a.TenantID == tenantID && a.ID == id {
			return a, nil
		}
	}
	return domain.AccessRequest{}, domain.ErrNotFound
}

func (r *AccessRepo) ListPending(_ context.Context, tenantID uuid.UUID) ([]domain.AccessRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AccessRequest{}
	for _, a := range r.s.Access {
		if a.TenantID == tenantID && a.Status == "pending" {
			out = append(out, a)
		}
	}
	return out, nil
}

type DeviceRepo struct{ s *Store }

func (r *DeviceRepo) Save(_ context.Context, d domain.DeviceTrust) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Devices {
		if r.s.Devices[i].ID == d.ID {
			r.s.Devices[i] = d
			return nil
		}
	}
	r.s.Devices = append(r.s.Devices, d)
	return nil
}

func (r *DeviceRepo) GetByDevice(_ context.Context, tenantID uuid.UUID, deviceID string) (domain.DeviceTrust, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, d := range r.s.Devices {
		if d.TenantID == tenantID && d.DeviceID == deviceID {
			return d, nil
		}
	}
	return domain.DeviceTrust{}, domain.ErrNotFound
}

type AISecRepo struct{ s *Store }

func (r *AISecRepo) Save(_ context.Context, e domain.AISecurityEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AISec = append(r.s.AISec, e)
	return nil
}

func (r *AISecRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.AISecurityEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AISecurityEvent{}
	for _, e := range r.s.AISec {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

type FraudRepo struct{ s *Store }

func (r *FraudRepo) Save(_ context.Context, s domain.FraudSignal) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Fraud = append(r.s.Fraud, s)
	return nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox = append(r.s.Outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Outbox {
		if r.s.Outbox[i].ID == m.ID {
			r.s.Outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

type MockVault struct{ Version int }

func (m *MockVault) Rotate(_ context.Context, _ string) (int, error) {
	m.Version++
	return m.Version, nil
}

func (m *MockVault) RenewCertificate(_ context.Context, _ string) (time.Time, error) {
	return time.Now().UTC().Add(90 * 24 * time.Hour), nil
}

type MockOPA struct{}

func (MockOPA) Evaluate(_ context.Context, rego string, input map[string]any) (bool, string, error) {
	_ = rego
	if action, _ := input["action"].(string); action == "deny_all" {
		return false, "explicit_deny", nil
	}
	if action, _ := input["action"].(string); action == "admin" {
		return false, "admin_requires_jit", nil
	}
	return true, "mock_allow", nil
}

type MockIdentity struct{}

func (MockIdentity) IdentityTrust(_ context.Context, _ uuid.UUID, _ string) (float64, error) {
	return 0.85, nil
}

type MockFraud struct{}

func (MockFraud) Score(_ context.Context, _ uuid.UUID, features map[string]float64) (float64, error) {
	var max float64
	for _, v := range features {
		if v > max {
			max = v
		}
	}
	if max > 1 {
		max = 1
	}
	return max, nil
}

type MockSIEM struct{ N int }

func (m *MockSIEM) SendAlert(_ context.Context, _ uuid.UUID, _ map[string]any) error {
	m.N++
	return nil
}

type MockSOAR struct{ N int }

func (m *MockSOAR) RunPlaybook(_ context.Context, _ uuid.UUID, playbookKey string, _ map[string]any) (string, error) {
	m.N++
	return "soar-" + playbookKey + "-" + uuid.NewString()[:8], nil
}

type MockAIGuard struct{}

func (MockAIGuard) ValidatePrompt(_ context.Context, _ uuid.UUID, _, prompt string) (bool, float64, error) {
	s := domain.PromptInjectionScore(prompt)
	return s >= 0.35, s, nil
}
