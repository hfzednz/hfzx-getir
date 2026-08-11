package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Policy kinds for OPA-backed decisions (not IAM role SoT).
const (
	PolicyAccess     = "access"
	PolicyData       = "data"
	PolicyCompliance = "compliance"
	PolicySecurity   = "security"
	PolicyApproval   = "approval"
	PolicyRuntime    = "runtime"
	PolicyAI         = "ai"
)

// SecurityPolicy is a versioned policy document.
type SecurityPolicy struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Kind      string
	Version   int
	Rego      string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ValidPolicyKind(k string) bool {
	switch k {
	case PolicyAccess, PolicyData, PolicyCompliance, PolicySecurity, PolicyApproval, PolicyRuntime, PolicyAI:
		return true
	default:
		return false
	}
}

// PolicyDecision is an evaluation result.
type PolicyDecision struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PolicyKey   string
	Subject     string
	Action      string
	Resource    string
	Allow       bool
	Reason      string
	RiskScore   float64
	Context     map[string]any
	EvaluatedAt time.Time
}

// AuditEvent is an immutable audit record (append-only).
type AuditEvent struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ActorID      string
	ActorType    string
	Action       string
	ResourceType string
	ResourceID   string
	Outcome      string
	IP           string
	UserAgent    string
	Metadata     map[string]any
	Hash         string
	PrevHash     string
	OccurredAt   time.Time
}

// SecretMeta is Vault-backed secret metadata (never stores raw secret).
type SecretMeta struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Kind        string
	VaultPath   string
	Version     int
	Rotatable   bool
	ExpiresAt   *time.Time
	LastRotated *time.Time
	Status      string
	CreatedAt   time.Time
}

// ThreatAlert anomaly / IDS signal.
type ThreatAlert struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Kind       string
	Severity   string
	Subject    string
	Score      float64
	Indicators map[string]any
	Status     string
	CreatedAt  time.Time
}

// ScanFinding DevSecOps / vuln finding.
type ScanFinding struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Source     string
	Target     string
	CVE        string
	Severity   string
	Title      string
	Status     string
	DetectedAt time.Time
	FixedAt    *time.Time
}

// Incident response case.
type Incident struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Title       string
	Severity    string
	Status      string
	ThreatID    *uuid.UUID
	Timeline    []IncidentEvent
	PlaybookKey string
	Assignee    string
	OpenedAt    time.Time
	ClosedAt    *time.Time
	Postmortem  string
}

type IncidentEvent struct {
	At      time.Time
	Actor   string
	Message string
}

// ComplianceControl control set.
type ComplianceControl struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Framework   string
	ControlID   string
	Title       string
	Status      string
	EvidenceIDs []uuid.UUID
	UpdatedAt   time.Time
}

type ComplianceEvidence struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ControlID   uuid.UUID
	Title       string
	URI         string
	CollectedAt time.Time
}

type ComplianceAuditRun struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Framework   string
	Score       float64
	Gaps        int
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
}

// DataAsset classification / PII.
type DataAsset struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Classification string
	PIITags        []string
	RetentionDays  int
	Owner          string
	CreatedAt      time.Time
}

type PrivacyRequest struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	SubjectRef  string
	Kind        string
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// RiskItem register entry.
type RiskItem struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Title      string
	Category   string
	Likelihood int
	Impact     int
	Score      int
	Status     string
	Owner      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func ComputeRiskScore(likelihood, impact int) int {
	if likelihood < 1 {
		likelihood = 1
	}
	if likelihood > 5 {
		likelihood = 5
	}
	if impact < 1 {
		impact = 1
	}
	if impact > 5 {
		impact = 5
	}
	return likelihood * impact
}

// AccessRequest temporary elevated access (JIT) — approval, not IAM role SoT.
type AccessRequest struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	RequesterID string
	RoleHint    string
	Resource    string
	Reason      string
	TTLMinutes  int
	Status      string
	CreatedAt   time.Time
	DecidedAt   *time.Time
	ExpiresAt   *time.Time
}

// DeviceTrust signal for Zero Trust.
type DeviceTrust struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	DeviceID   string
	Platform   string
	Attested   bool
	Rooted     bool
	Jailbroken bool
	Tampered   bool
	TrustScore float64
	LastSeenAt time.Time
}

// AISecurityEvent prompt injection / guardrail.
type AISecurityEvent struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ModelKey   string
	PromptHash string
	Kind       string
	Blocked    bool
	Score      float64
	CreatedAt  time.Time
}

// FraudSignal behavioral — may forward to fraud-service facade.
type FraudSignal struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Subject   string
	Kind      string
	Score     float64
	Features  map[string]float64
	CreatedAt time.Time
}

func NormalizeKey(k string) string {
	return strings.ToLower(strings.TrimSpace(k))
}

func ValidSeverity(s string) bool {
	switch s {
	case "low", "medium", "high", "critical":
		return true
	default:
		return false
	}
}

func ValidClassification(c string) bool {
	switch c {
	case "public", "internal", "confidential", "restricted":
		return true
	default:
		return false
	}
}

func ChainHash(prev, payload string) string {
	sum := uint32(2166136261)
	for _, b := range []byte(prev + "|" + payload) {
		sum ^= uint32(b)
		sum *= 16777619
	}
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		out[i] = hexdigits[sum&0xf]
		sum >>= 4
	}
	return string(out)
}

func AdaptiveTrustScore(identityTrust, deviceTrust, contextRisk float64) float64 {
	if identityTrust < 0 {
		identityTrust = 0
	}
	if identityTrust > 1 {
		identityTrust = 1
	}
	if deviceTrust < 0 {
		deviceTrust = 0
	}
	if deviceTrust > 1 {
		deviceTrust = 1
	}
	if contextRisk < 0 {
		contextRisk = 0
	}
	if contextRisk > 1 {
		contextRisk = 1
	}
	score := 0.45*identityTrust + 0.35*deviceTrust + 0.20*(1 - contextRisk)
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func DetectThreatKind(features map[string]float64) (kind string, score float64) {
	score = 0
	kind = "anomaly"
	if features["failed_logins"] > 10 {
		return "brute_force", clamp01(features["failed_logins"] / 50)
	}
	if features["new_device"] > 0 && features["geo_distance_km"] > 500 {
		return "ato", clamp01(0.6 + features["geo_distance_km"]/5000)
	}
	if features["request_rate"] > 100 {
		return "bot", clamp01(features["request_rate"] / 500)
	}
	if features["password_spray"] > 0 {
		return "credential_stuffing", 0.8
	}
	if features["role_change"] > 0 && features["unusual_hour"] > 0 {
		return "priv_esc", 0.75
	}
	for _, v := range features {
		if v > score {
			score = v
		}
	}
	return kind, clamp01(score)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func ComplianceScore(controls []ComplianceControl) (score float64, gaps int) {
	if len(controls) == 0 {
		return 0, 0
	}
	met := 0
	for _, c := range controls {
		if c.Status == "met" {
			met++
		} else if c.Status == "gap" || c.Status == "not_started" {
			gaps++
		}
	}
	return (float64(met) / float64(len(controls))) * 100, gaps
}

func PromptInjectionScore(prompt string) float64 {
	p := strings.ToLower(prompt)
	hits := 0
	for _, needle := range []string{"ignore previous", "system prompt", "jailbreak", "exfiltrate", "bypass safety"} {
		if strings.Contains(p, needle) {
			hits++
		}
	}
	return clamp01(float64(hits) * 0.35)
}

func SeverityFromScore(score float64) string {
	switch {
	case score >= 0.85:
		return "critical"
	case score >= 0.65:
		return "high"
	case score >= 0.4:
		return "medium"
	default:
		return "low"
	}
}
