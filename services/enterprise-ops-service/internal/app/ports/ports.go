package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type OrgRepo interface {
	Save(ctx context.Context, n domain.OrgNode) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.OrgNode, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.OrgNode, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.OrgNode, error)
	ListChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]domain.OrgNode, error)
}

type PolicyRepo interface {
	Save(ctx context.Context, p domain.Policy) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Policy, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Policy, error)
}

type PortfolioRepo interface {
	Save(ctx context.Context, p domain.Portfolio) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Portfolio, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Portfolio, error)
}

type ProgramRepo interface {
	Save(ctx context.Context, p domain.Program) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Program, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Program, error)
}

type ProjectRepo interface {
	Save(ctx context.Context, p domain.Project) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Project, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Project, error)
}

type MilestoneRepo interface {
	Save(ctx context.Context, m domain.Milestone) error
	ListByProject(ctx context.Context, tenantID, projectID uuid.UUID) ([]domain.Milestone, error)
}

type ObjectiveRepo interface {
	Save(ctx context.Context, o domain.Objective) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Objective, error)
}

type KeyResultRepo interface {
	Save(ctx context.Context, kr domain.KeyResult) error
	ListByObjective(ctx context.Context, tenantID, objectiveID uuid.UUID) ([]domain.KeyResult, error)
}

type KPIRepo interface {
	Save(ctx context.Context, k domain.KPI) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.KPI, error)
}

type RiskRepo interface {
	Save(ctx context.Context, r domain.Risk) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Risk, error)
}

type ContinuityRepo interface {
	Save(ctx context.Context, p domain.ContinuityPlan) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.ContinuityPlan, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ContinuityPlan, error)
}

type AuditRepo interface {
	Save(ctx context.Context, a domain.AuditEngagement) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.AuditEngagement, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.AuditEngagement, error)
}

type FindingRepo interface {
	Save(ctx context.Context, f domain.AuditFinding) error
	ListByAudit(ctx context.Context, tenantID, auditID uuid.UUID) ([]domain.AuditFinding, error)
}

type MeetingRepo interface {
	Save(ctx context.Context, m domain.Meeting) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Meeting, error)
}

type DecisionRepo interface {
	Save(ctx context.Context, d domain.Decision) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Decision, error)
}

type KnowledgeRepo interface {
	Save(ctx context.Context, d domain.KnowledgeDoc) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.KnowledgeDoc, error)
}

type ResourceRepo interface {
	Save(ctx context.Context, r domain.ResourcePlan) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ResourcePlan, error)
}

// SecurityClient validates policy approval against security GRC (port only).
type SecurityClient interface {
	PolicyChangeAllowed(ctx context.Context, tenantID uuid.UUID, policyKey string) (bool, error)
}

// AIClient executive insights / forecasting.
type AIClient interface {
	ProjectForecast(ctx context.Context, tenantID uuid.UUID, projectCode string) (map[string]any, error)
	RiskPrediction(ctx context.Context, tenantID uuid.UUID) ([]string, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
