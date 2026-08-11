package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type FlagRepo interface {
	Save(ctx context.Context, f domain.FeatureFlag) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.FeatureFlag, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.FeatureFlag, error)
}

type ConfigRepo interface {
	Save(ctx context.Context, c domain.ConfigDocument) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.ConfigDocument, error)
	List(ctx context.Context, tenantID uuid.UUID, namespace string) ([]domain.ConfigDocument, error)
}

type ExperimentRepo interface {
	Save(ctx context.Context, e domain.Experiment) error
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Experiment, error)
	SaveAssignment(ctx context.Context, a domain.Assignment) error
	GetAssignment(ctx context.Context, tenantID, experimentID uuid.UUID, subjectID string) (domain.Assignment, bool, error)
}

type EventRepo interface {
	Save(ctx context.Context, e domain.LiveOpsEvent) error
	List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.LiveOpsEvent, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.LiveOpsEvent, error)
}

type ChangeRepo interface {
	Save(ctx context.Context, c domain.ChangeRequest) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error)
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.ChangeRequest, error)
}

type RollbackRepo interface {
	Save(ctx context.Context, r domain.RollbackRecord) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RollbackRecord, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type EvalCache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

// MetricsClient sends experiment metrics to data-platform (not SoT here).
type MetricsClient interface {
	Ingest(ctx context.Context, tenantID uuid.UUID, metric string, value float64, tags map[string]string) error
}

// AIClient optional winner / experiment suggestions.
type AIClient interface {
	SuggestWinner(ctx context.Context, tenantID uuid.UUID, experimentKey string, rates map[string]float64) (string, error)
}
