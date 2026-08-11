package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
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
	Save(ctx context.Context, a domain.Audit) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Audit, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Audit, error)
}

type FindingRepo interface {
	Save(ctx context.Context, f domain.Finding) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Finding, error)
	ListByAudit(ctx context.Context, tenantID, auditID uuid.UUID) ([]domain.Finding, error)
	OpenCritical(ctx context.Context, tenantID uuid.UUID) (int, error)
}

type BenchmarkRepo interface {
	Save(ctx context.Context, b domain.BenchmarkRun) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.BenchmarkRun, error)
	LatestByKind(ctx context.Context, tenantID uuid.UUID, kind domain.BenchmarkKind) (domain.BenchmarkRun, error)
}

type CapacityRepo interface {
	Save(ctx context.Context, c domain.CapacityScenario) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.CapacityScenario, error)
}

type ChaosRepo interface {
	Save(ctx context.Context, c domain.ChaosExperiment) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChaosExperiment, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ChaosExperiment, error)
}

type TuningRepo interface {
	Save(ctx context.Context, t domain.TuningProfile) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.TuningProfile, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.TuningProfile, error)
}

type CertificateRepo interface {
	Save(ctx context.Context, c domain.Certificate) error
	Latest(ctx context.Context, tenantID uuid.UUID) (domain.Certificate, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Certificate, error)
}

type QualityClient interface {
	ReleaseGatesGreen(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type PlatformOpsClient interface {
	DRDrillPassed(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type SecurityClient interface {
	ZeroCriticalVulns(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
