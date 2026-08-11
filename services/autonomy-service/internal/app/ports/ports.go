package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/domain"
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

type AuditRepo interface {
	Save(ctx context.Context, a domain.AutonomyAudit) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.AutonomyAudit, error)
}

type WeaknessRepo interface {
	Save(ctx context.Context, w domain.Weakness) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Weakness, error)
	OpenCount(ctx context.Context, tenantID uuid.UUID) (int, error)
}

type HealRepo interface {
	Save(ctx context.Context, a domain.HealAction) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.HealAction, error)
	ExecutedCount(ctx context.Context, tenantID uuid.UUID) (int, error)
}

type ReviewRepo interface {
	Save(ctx context.Context, r domain.AICTOReview) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.AICTOReview, error)
}

type EvolutionRepo interface {
	Save(ctx context.Context, t domain.EvolutionTask) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.EvolutionTask, error)
}

type ReleaseRepo interface {
	Save(ctx context.Context, p domain.ReleasePlan) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ReleasePlan, error)
}

type GovernanceRepo interface {
	Save(ctx context.Context, g domain.GovernanceLoop) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GovernanceLoop, error)
}

type AssistantRepo interface {
	Save(ctx context.Context, a domain.ExecutiveAssistant) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ExecutiveAssistant, error)
}

type TeamRepo interface {
	Save(ctx context.Context, t domain.DigitalTeam) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DigitalTeam, error)
}

type DependencyRepo interface {
	Save(ctx context.Context, e domain.DependencyEdge) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DependencyEdge, error)
}

type GenesisRepo interface {
	Save(ctx context.Context, c domain.GenesisCertificate) error
	Latest(ctx context.Context, tenantID uuid.UUID) (domain.GenesisCertificate, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GenesisCertificate, error)
}

type HyperscaleClient interface {
	Certified(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type PlatformOpsClient interface {
	ExecuteHeal(ctx context.Context, tenantID uuid.UUID, action, target string) error
}

type QualityClient interface {
	Healthy(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type SecurityClient interface {
	Healthy(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type LiveOpsClient interface {
	EmergencyOffClear(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
