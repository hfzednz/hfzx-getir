package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuditScope string

const (
	ScopeDependency   AuditScope = "dependency"
	ScopeArchitecture AuditScope = "architecture"
	ScopeBusiness     AuditScope = "business"
	ScopeInfrastructure AuditScope = "infrastructure"
	ScopeSecurity     AuditScope = "security"
	ScopePerformance  AuditScope = "performance"
	ScopeAI           AuditScope = "ai"
	ScopeDocumentation AuditScope = "documentation"
	ScopeCompliance   AuditScope = "compliance"
	ScopeDX           AuditScope = "developer_experience"
	ScopeOperational  AuditScope = "operational"
)

type AutonomyAudit struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenantId"`
	Scope       AuditScope `json:"scope"`
	Status      string     `json:"status"` // running|completed
	Score       float64    `json:"score"`  // 0..100
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type Weakness struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	AuditID    uuid.UUID `json:"auditId"`
	Code       string    `json:"code"`
	Title      string    `json:"title"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"` // open|resolved
	Resolution string    `json:"resolution"`
	CreatedAt  time.Time `json:"createdAt"`
}

type HealKind string

const (
	HealService  HealKind = "service"
	HealDatabase HealKind = "database"
	HealInfra    HealKind = "infrastructure"
	HealSecurity HealKind = "security"
	HealData     HealKind = "data"
	HealAI       HealKind = "ai"
	HealQA       HealKind = "qa"
)

type HealAction struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Kind        HealKind  `json:"kind"`
	TargetRef   string    `json:"targetRef"` // opaque service/resource
	Action      string    `json:"action"`    // restart|rollback|scale|rotate|retrain|patch
	Status      string    `json:"status"`    // planned|executed|failed
	Automated   bool      `json:"automated"`
	CreatedAt   time.Time `json:"createdAt"`
	ExecutedAt  *time.Time `json:"executedAt,omitempty"`
}

type ReviewKind string

const (
	ReviewArchitecture ReviewKind = "architecture"
	ReviewDependency   ReviewKind = "dependency"
	ReviewSecurity     ReviewKind = "security"
	ReviewPerformance  ReviewKind = "performance"
	ReviewCost         ReviewKind = "cost"
	ReviewDatabase     ReviewKind = "database"
	ReviewInfrastructure ReviewKind = "infrastructure"
	ReviewRelease      ReviewKind = "release"
)

type AICTOReview struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenantId"`
	Kind         ReviewKind `json:"kind"`
	Summary      string     `json:"summary"`
	DebtScore    float64    `json:"debtScore"` // 0..100 lower better
	Suggestions  []string   `json:"suggestions"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type EvolutionKind string

const (
	EvoRefactor    EvolutionKind = "refactor"
	EvoDeadCode    EvolutionKind = "dead_code"
	EvoDependency  EvolutionKind = "dependency_upgrade"
	EvoAPI         EvolutionKind = "api_evolution"
	EvoSchema      EvolutionKind = "schema_evolution"
	EvoDocs        EvolutionKind = "documentation"
	EvoDebt        EvolutionKind = "tech_debt"
)

type EvolutionTask struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenantId"`
	Kind      EvolutionKind `json:"kind"`
	Title     string        `json:"title"`
	Priority  int           `json:"priority"` // 1 highest
	Status    string        `json:"status"`   // backlog|in_progress|done
	CreatedAt time.Time     `json:"createdAt"`
}

type ReleasePlan struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Version   string    `json:"version"`
	Strategy  string    `json:"strategy"` // canary|blue_green
	Score     float64   `json:"score"`    // 0..100
	Status    string    `json:"status"`   // planned|validated|rolled_out|rolled_back
	CreatedAt time.Time `json:"createdAt"`
}

type GovernanceLoop struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Domain    string    `json:"domain"` // architecture|security|ai|data|operational|compliance
	Cadence   string    `json:"cadence"` // continuous|daily|weekly
	Healthy   bool      `json:"healthy"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AssistantRole string

const (
	AssistCEO AssistantRole = "CEO"
	AssistCOO AssistantRole = "COO"
	AssistCTO AssistantRole = "CTO"
	AssistCFO AssistantRole = "CFO"
	AssistCMO AssistantRole = "CMO"
	AssistOps AssistantRole = "Operations"
	AssistFin AssistantRole = "Finance"
	AssistSec AssistantRole = "Security"
)

type ExecutiveAssistant struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenantId"`
	Role      AssistantRole `json:"role"`
	Name      string        `json:"name"`
	Active    bool          `json:"active"`
	CreatedAt time.Time     `json:"createdAt"`
}

type DigitalTeamKind string

const (
	TeamEngineering DigitalTeamKind = "engineering"
	TeamOperations  DigitalTeamKind = "operations"
	TeamSupport     DigitalTeamKind = "support"
	TeamAnalysts    DigitalTeamKind = "analysts"
	TeamReviewers   DigitalTeamKind = "reviewers"
	TeamPlanners    DigitalTeamKind = "planners"
	TeamArchitects  DigitalTeamKind = "architects"
)

type DigitalTeam struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenantId"`
	Kind      DigitalTeamKind `json:"kind"`
	Name      string          `json:"name"`
	Active    bool            `json:"active"`
	CreatedAt time.Time       `json:"createdAt"`
}

type DependencyEdge struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	FromService string   `json:"fromService"`
	ToService   string   `json:"toService"`
	Relation    string   `json:"relation"` // http|grpc|kafka|port
	CreatedAt   time.Time `json:"createdAt"`
}

type GenesisCertificate struct {
	ID        uuid.UUID         `json:"id"`
	TenantID  uuid.UUID         `json:"tenantId"`
	Version   string            `json:"version"`
	Status    string            `json:"status"` // pending|issued
	Gates     map[string]bool   `json:"gates"`
	IssuedAt  *time.Time        `json:"issuedAt,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

func GenesisGatesRequired() []string {
	return []string{
		"autonomy_audits", "self_healing", "ai_cto", "evolution",
		"release_engine", "governance", "hyperscale", "security", "quality",
	}
}

func ReleaseScore(canaryOK, testsOK, securityOK bool, errorBudgetBurn float64) float64 {
	score := 40.0
	if canaryOK {
		score += 20
	}
	if testsOK {
		score += 20
	}
	if securityOK {
		score += 15
	}
	if errorBudgetBurn < 0.5 {
		score += 5
	}
	if score > 100 {
		return 100
	}
	return score
}

func ValidateHeal(a HealAction) error {
	if a.TenantID == uuid.Nil || a.Kind == "" || strings.TrimSpace(a.TargetRef) == "" || a.Action == "" {
		return ErrInvalidArgument
	}
	return nil
}

func DefaultDependencyGraph() []DependencyEdge {
	edges := []struct{ from, to, rel string }{
		{"bff-customer", "order-service", "http"},
		{"bff-customer", "catalog-service", "http"},
		{"bff-customer", "payment-service", "http"},
		{"order-service", "inventory-service", "http"},
		{"order-service", "payment-service", "http"},
		{"dispatch-service", "tracking-service", "kafka"},
		{"liveops-service", "bff-customer", "port"},
		{"superapp-service", "liveops-service", "port"},
		{"innovation-service", "liveops-service", "port"},
		{"enterprise-ops-service", "security-service", "port"},
		{"hyperscale-cert-service", "quality-service", "port"},
		{"hyperscale-cert-service", "platform-ops-service", "port"},
		{"autonomy-service", "hyperscale-cert-service", "port"},
		{"autonomy-service", "platform-ops-service", "port"},
		{"autonomy-service", "quality-service", "port"},
		{"autonomy-service", "security-service", "port"},
		{"autonomy-service", "liveops-service", "port"},
	}
	out := make([]DependencyEdge, 0, len(edges))
	for _, e := range edges {
		out = append(out, DependencyEdge{FromService: e.from, ToService: e.to, Relation: e.rel})
	}
	return out
}

func DefaultAssistants() []ExecutiveAssistant {
	roles := []AssistantRole{AssistCEO, AssistCOO, AssistCTO, AssistCFO, AssistCMO, AssistOps, AssistFin, AssistSec}
	out := make([]ExecutiveAssistant, 0, len(roles))
	for _, r := range roles {
		out = append(out, ExecutiveAssistant{Role: r, Name: string(r) + " Assistant", Active: true})
	}
	return out
}

func DefaultDigitalTeams() []DigitalTeam {
	kinds := []DigitalTeamKind{TeamEngineering, TeamOperations, TeamSupport, TeamAnalysts, TeamReviewers, TeamPlanners, TeamArchitects}
	out := make([]DigitalTeam, 0, len(kinds))
	for _, k := range kinds {
		out = append(out, DigitalTeam{Kind: k, Name: "AI " + string(k), Active: true})
	}
	return out
}
