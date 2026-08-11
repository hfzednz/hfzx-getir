package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/domain"
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

type ModuleRepo interface {
	Save(ctx context.Context, m domain.Module) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Module, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Module, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Module, error)
}

type ManifestRepo interface {
	Save(ctx context.Context, m domain.ModuleManifest) error
	Get(ctx context.Context, tenantID, moduleID uuid.UUID, version string) (domain.ModuleManifest, error)
	Latest(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.ModuleManifest, error)
}

type InstallRepo interface {
	Save(ctx context.Context, i domain.Install) error
	Get(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) (domain.Install, error)
	ListBySubject(ctx context.Context, tenantID uuid.UUID, subjectID string) ([]domain.Install, error)
}

type PermissionRepo interface {
	Save(ctx context.Context, g domain.PermissionGrant) error
	List(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID) ([]domain.PermissionGrant, error)
	Has(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID, perm string) (bool, error)
}

type ListingRepo interface {
	Save(ctx context.Context, l domain.StoreListing) error
	GetByModule(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.StoreListing, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.StoreListing, error)
}

type RatingRepo interface {
	Save(ctx context.Context, r domain.StoreRating) error
	ListByModule(ctx context.Context, tenantID, moduleID uuid.UUID) ([]domain.StoreRating, error)
}

type WidgetRepo interface {
	Save(ctx context.Context, w domain.WidgetPlacement) error
	ListBySubject(ctx context.Context, tenantID uuid.UUID, subjectID string) ([]domain.WidgetPlacement, error)
}

type MonetizationRepo interface {
	Save(ctx context.Context, r domain.MonetizationRule) error
	GetByModule(ctx context.Context, tenantID, moduleID uuid.UUID) (domain.MonetizationRule, error)
}

type LaunchRepo interface {
	Save(ctx context.Context, e domain.LaunchEvent) error
}

type LiveOpsClient interface {
	ModuleEnabled(ctx context.Context, tenantID uuid.UUID, moduleKey string, subjectID string) (bool, error)
}

type AIClient interface {
	RecommendModules(ctx context.Context, tenantID uuid.UUID, subjectID string, limit int) ([]string, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
