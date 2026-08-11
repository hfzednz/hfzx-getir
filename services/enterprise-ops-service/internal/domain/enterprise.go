package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrgNodeKind string

const (
	OrgCompany    OrgNodeKind = "company"
	OrgBusinessUnit OrgNodeKind = "business_unit"
	OrgDepartment OrgNodeKind = "department"
	OrgTeam       OrgNodeKind = "team"
	OrgBranch     OrgNodeKind = "branch"
	OrgRegion     OrgNodeKind = "region"
)

type OrgNode struct {
	ID           uuid.UUID   `json:"id"`
	TenantID     uuid.UUID   `json:"tenantId"`
	Kind         OrgNodeKind `json:"kind"`
	Code         string      `json:"code"`
	Name         string      `json:"name"`
	ParentID     *uuid.UUID  `json:"parentId,omitempty"`
	ManagerRef   string      `json:"managerRef"` // opaque identity subject
	CountryCode  string      `json:"countryCode"`
	Active       bool        `json:"active"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type PolicyStatus string

const (
	PolicyDraft     PolicyStatus = "draft"
	PolicyInReview  PolicyStatus = "in_review"
	PolicyApproved  PolicyStatus = "approved"
	PolicyRetired   PolicyStatus = "retired"
)

type PolicyKind string

const (
	PolicyCorporate  PolicyKind = "corporate"
	PolicyOperational PolicyKind = "operational"
	PolicySecurity   PolicyKind = "security" // metadata only; enforcement in security-service
	PolicyLegal      PolicyKind = "legal"
)

type Policy struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenantId"`
	Key         string       `json:"key"`
	Title       string       `json:"title"`
	Kind        PolicyKind   `json:"kind"`
	Status      PolicyStatus `json:"status"`
	Version     string       `json:"version"`
	BodyURI     string       `json:"bodyUri"`
	OwnerRef    string       `json:"ownerRef"`
	ApprovedBy  string       `json:"approvedBy,omitempty"`
	ApprovedAt  *time.Time   `json:"approvedAt,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type ProjectStatus string

const (
	ProjectPlanned   ProjectStatus = "planned"
	ProjectActive    ProjectStatus = "active"
	ProjectOnHold    ProjectStatus = "on_hold"
	ProjectCompleted ProjectStatus = "completed"
	ProjectCancelled ProjectStatus = "cancelled"
)

type Portfolio struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	OwnerRef  string    `json:"ownerRef"`
	CreatedAt time.Time `json:"createdAt"`
}

type Program struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	PortfolioID uuid.UUID `json:"portfolioId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Project struct {
	ID            uuid.UUID     `json:"id"`
	TenantID      uuid.UUID     `json:"tenantId"`
	ProgramID     uuid.UUID     `json:"programId"`
	Code          string        `json:"code"`
	Name          string        `json:"name"`
	Status        ProjectStatus `json:"status"`
	BudgetMinor   int64         `json:"budgetMinor"`
	Currency      string        `json:"currency"`
	Health        string        `json:"health"` // green|amber|red
	OwnerRef      string        `json:"ownerRef"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type Milestone struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	ProjectID uuid.UUID `json:"projectId"`
	Name      string    `json:"name"`
	DueAt     time.Time `json:"dueAt"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
}

type Objective struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Period    string    `json:"period"` // YYYY-Qn or YYYY
	Title     string    `json:"title"`
	OwnerRef  string    `json:"ownerRef"`
	CreatedAt time.Time `json:"createdAt"`
}

type KeyResult struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	ObjectiveID uuid.UUID `json:"objectiveId"`
	Title       string    `json:"title"`
	Target      float64   `json:"target"`
	Current     float64   `json:"current"`
	Unit        string    `json:"unit"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KPI struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Target    float64   `json:"target"`
	Period    string    `json:"period"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type RiskCategory string

const (
	RiskEnterprise  RiskCategory = "enterprise"
	RiskOperational RiskCategory = "operational"
	RiskFinancial   RiskCategory = "financial"
	RiskTechnology  RiskCategory = "technology"
	RiskCompliance  RiskCategory = "compliance"
	RiskStrategic   RiskCategory = "strategic"
)

type Risk struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenantId"`
	Code        string       `json:"code"`
	Title       string       `json:"title"`
	Category    RiskCategory `json:"category"`
	Likelihood  int          `json:"likelihood"` // 1..5
	Impact      int          `json:"impact"`     // 1..5
	Score       int          `json:"score"`
	Status      string       `json:"status"` // open|mitigating|closed
	OwnerRef    string       `json:"ownerRef"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type ContinuityPlan struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	RTOHours     int       `json:"rtoHours"`
	RPOHours     int       `json:"rpoHours"`
	Priority     int       `json:"priority"` // 1 highest
	CriticalSvc  string    `json:"criticalService"`
	Status       string    `json:"status"` // draft|approved|active|retired
	ActivatedAt  *time.Time `json:"activatedAt,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AuditEngagement struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Kind        string    `json:"kind"` // internal|external
	Status      string    `json:"status"` // planned|in_progress|completed
	ScheduledAt time.Time `json:"scheduledAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AuditFinding struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	AuditID    uuid.UUID `json:"auditId"`
	Severity   string    `json:"severity"` // low|medium|high|critical
	Title      string    `json:"title"`
	CAPA       string    `json:"capa"`
	Status     string    `json:"status"` // open|remediating|closed
	CreatedAt  time.Time `json:"createdAt"`
}

type MeetingKind string

const (
	MeetingBoard      MeetingKind = "board"
	MeetingExecutive  MeetingKind = "executive"
	MeetingOperational MeetingKind = "operational"
)

type Meeting struct {
	ID         uuid.UUID   `json:"id"`
	TenantID   uuid.UUID   `json:"tenantId"`
	Kind       MeetingKind `json:"kind"`
	Title      string      `json:"title"`
	StartsAt   time.Time   `json:"startsAt"`
	Agenda     string      `json:"agenda"`
	MinutesURI string      `json:"minutesUri,omitempty"`
	Status     string      `json:"status"` // scheduled|completed|cancelled
	CreatedAt  time.Time   `json:"createdAt"`
}

type Decision struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	MeetingID  *uuid.UUID `json:"meetingId,omitempty"`
	DecidedBy  string    `json:"decidedBy"`
	VotesFor   int       `json:"votesFor"`
	VotesAgainst int     `json:"votesAgainst"`
	Impact     string    `json:"impact"`
	CreatedAt  time.Time `json:"createdAt"`
}

type KnowledgeDoc struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"` // wiki|playbook|runbook|best_practice
	URI       string    `json:"uri"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResourcePlan struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	TeamCode     string    `json:"teamCode"`
	Period       string    `json:"period"`
	CapacityFTE  float64   `json:"capacityFte"`
	AllocatedFTE float64   `json:"allocatedFte"`
	Utilization  float64   `json:"utilization"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func ValidateOrg(n OrgNode) error {
	if n.TenantID == uuid.Nil || n.Kind == "" || strings.TrimSpace(n.Code) == "" || n.Name == "" {
		return ErrInvalidArgument
	}
	return nil
}

func ValidatePolicy(p Policy) error {
	if p.TenantID == uuid.Nil || p.Key == "" || p.Title == "" || p.Kind == "" {
		return ErrInvalidArgument
	}
	return nil
}

func RiskScore(likelihood, impact int) (int, error) {
	if likelihood < 1 || likelihood > 5 || impact < 1 || impact > 5 {
		return 0, ErrInvalidArgument
	}
	return likelihood * impact, nil
}

func ProjectHealth(budgetUsedPct float64, openRisks int) string {
	if openRisks >= 3 || budgetUsedPct > 1.1 {
		return "red"
	}
	if openRisks >= 1 || budgetUsedPct > 0.9 {
		return "amber"
	}
	return "green"
}
