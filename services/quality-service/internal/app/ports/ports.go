package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/domain"
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

type SuiteRepo interface {
	Save(ctx context.Context, s domain.Suite) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Suite, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Suite, error)
}

type RunRepo interface {
	Save(ctx context.Context, r domain.TestRun) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.TestRun, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.TestRun, error)
}

type ResultRepo interface {
	Save(ctx context.Context, r domain.TestCaseResult) error
	ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.TestCaseResult, error)
}

type CoverageRepo interface {
	Save(ctx context.Context, c domain.CoverageReport) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.CoverageReport, error)
}

type PolicyRepo interface {
	Save(ctx context.Context, p domain.GatePolicy) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GatePolicy, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.GatePolicy, error)
}

type EvalRepo interface {
	Save(ctx context.Context, e domain.GateEvaluation) error
	ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.GateEvaluation, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GateEvaluation, error)
}

type CertRepo interface {
	Save(ctx context.Context, c domain.Certification) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Certification, error)
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Certification, error)
}

type FlakyRepo interface {
	Save(ctx context.Context, f domain.FlakyRecord) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.FlakyRecord, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, suiteKey, name string) (domain.FlakyRecord, error)
}

type PerfRepo interface {
	Save(ctx context.Context, p domain.PerfMetric) error
	ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.PerfMetric, error)
}

type SecurityRepo interface {
	Save(ctx context.Context, f domain.SecurityFinding) error
	ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.SecurityFinding, error)
}

type RunnerClient interface {
	Dispatch(ctx context.Context, suite domain.Suite, run domain.TestRun) error
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
