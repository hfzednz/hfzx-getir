package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/domain"
)

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// IDGen generates opaque UUIDs.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, msg domain.OutboxMessage) error
	Update(ctx context.Context, msg domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// ZoneRepo stores geofence zones.
type ZoneRepo interface {
	Create(ctx context.Context, z domain.Zone) error
	Update(ctx context.Context, z domain.Zone) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Zone, error)
	List(ctx context.Context, tenantID uuid.UUID, city string, kind domain.ZoneKind, limit, offset int) ([]domain.Zone, int, error)
	ListActive(ctx context.Context, tenantID uuid.UUID, city string) ([]domain.Zone, error)
}
