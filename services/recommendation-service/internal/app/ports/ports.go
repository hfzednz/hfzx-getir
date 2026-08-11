package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type FeatureRepo interface {
	Upsert(ctx context.Context, f domain.ProductFeatures) error
	Get(ctx context.Context, productID uuid.UUID) (domain.ProductFeatures, error)
	List(ctx context.Context, tenantScoped []uuid.UUID) ([]domain.ProductFeatures, error)
	ListAll(ctx context.Context, limit int) ([]domain.ProductFeatures, error)
}

type SignalRepo interface {
	Save(ctx context.Context, s domain.BehaviorSignal) error
	ListByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.BehaviorSignal, error)
	UsersWhoInteracted(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]uuid.UUID, error)
}

type CoOccurRepo interface {
	Bump(ctx context.Context, tenantID, a, b uuid.UUID, delta int, now time.Time) error
	TopFor(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.CoOccurrence, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type CatalogClient interface {
	Features(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductFeatures, error)
}

type TrendClient interface {
	TrendingProductIDs(ctx context.Context, tenantID uuid.UUID, limit int) ([]uuid.UUID, error)
}
