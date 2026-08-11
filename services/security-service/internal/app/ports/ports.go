package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type PolicyRepo interface {
	Save(ctx context.Context, p domain.SecurityPolicy) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SecurityPolicy, error)
	List(ctx context.Context, tenantID uuid.UUID, kind string) ([]domain.SecurityPolicy, error)
	SaveDecision(ctx context.Context, d domain.PolicyDecision) error
}

type AuditRepo interface {
	Append(ctx context.Context, e domain.AuditEvent) error
	LastHash(ctx context.Context, tenantID uuid.UUID) (string, error)
	Search(ctx context.Context, tenantID uuid.UUID, action, actor string, limit int) ([]domain.AuditEvent, error)
}

type SecretRepo interface {
	Save(ctx context.Context, s domain.SecretMeta) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SecretMeta, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.SecretMeta, error)
}

type ThreatRepo interface {
	Save(ctx context.Context, t domain.ThreatAlert) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ThreatAlert, error)
	List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.ThreatAlert, error)
}

type VulnRepo interface {
	Save(ctx context.Context, f domain.ScanFinding) error
	List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.ScanFinding, error)
}

type IncidentRepo interface {
	Save(ctx context.Context, i domain.Incident) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Incident, error)
	List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.Incident, error)
}

type ComplianceRepo interface {
	SaveControl(ctx context.Context, c domain.ComplianceControl) error
	ListControls(ctx context.Context, tenantID uuid.UUID, framework string) ([]domain.ComplianceControl, error)
	SaveEvidence(ctx context.Context, e domain.ComplianceEvidence) error
	SaveAuditRun(ctx context.Context, r domain.ComplianceAuditRun) error
	ListAuditRuns(ctx context.Context, tenantID uuid.UUID) ([]domain.ComplianceAuditRun, error)
}

type DataGovRepo interface {
	SaveAsset(ctx context.Context, a domain.DataAsset) error
	ListAssets(ctx context.Context, tenantID uuid.UUID) ([]domain.DataAsset, error)
	SavePrivacy(ctx context.Context, p domain.PrivacyRequest) error
	GetPrivacy(ctx context.Context, tenantID, id uuid.UUID) (domain.PrivacyRequest, error)
	ListPrivacy(ctx context.Context, tenantID uuid.UUID) ([]domain.PrivacyRequest, error)
}

type RiskRepo interface {
	Save(ctx context.Context, r domain.RiskItem) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.RiskItem, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RiskItem, error)
}

type AccessRepo interface {
	Save(ctx context.Context, a domain.AccessRequest) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.AccessRequest, error)
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.AccessRequest, error)
}

type DeviceRepo interface {
	Save(ctx context.Context, d domain.DeviceTrust) error
	GetByDevice(ctx context.Context, tenantID uuid.UUID, deviceID string) (domain.DeviceTrust, error)
}

type AISecRepo interface {
	Save(ctx context.Context, e domain.AISecurityEvent) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.AISecurityEvent, error)
}

type FraudRepo interface {
	Save(ctx context.Context, s domain.FraudSignal) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// VaultClient rotates/reads metadata only — never returns plaintext to callers here.
type VaultClient interface {
	Rotate(ctx context.Context, path string) (newVersion int, err error)
	RenewCertificate(ctx context.Context, path string) (expiresAt time.Time, err error)
}

// OPAClient evaluates Rego policies.
type OPAClient interface {
	Evaluate(ctx context.Context, rego string, input map[string]any) (allow bool, reason string, err error)
}

// IdentityTrustClient reads trust signals from IAM (no auth ownership).
type IdentityTrustClient interface {
	IdentityTrust(ctx context.Context, tenantID uuid.UUID, subject string) (float64, error)
}

// FraudFacadeClient optional fraud-service scoring.
type FraudFacadeClient interface {
	Score(ctx context.Context, tenantID uuid.UUID, features map[string]float64) (float64, error)
}

// SIEMClient forwards alerts.
type SIEMClient interface {
	SendAlert(ctx context.Context, tenantID uuid.UUID, alert map[string]any) error
}

// SOARClient triggers playbooks.
type SOARClient interface {
	RunPlaybook(ctx context.Context, tenantID uuid.UUID, playbookKey string, context map[string]any) (runID string, err error)
}

// AIGuardrailClient optional ai-platform check.
type AIGuardrailClient interface {
	ValidatePrompt(ctx context.Context, tenantID uuid.UUID, modelKey, prompt string) (blocked bool, score float64, err error)
}
