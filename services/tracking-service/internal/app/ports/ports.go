// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// LocationRepo persists latest courier locations and capped history.
type LocationRepo interface {
	UpsertLatest(ctx context.Context, loc domain.CourierLocation) error
	GetLatest(ctx context.Context, tenantID, courierID uuid.UUID) (domain.CourierLocation, error)
	AppendHistory(ctx context.Context, entry domain.LocationHistoryEntry, cap int) error
	ListHistory(ctx context.Context, tenantID, courierID uuid.UUID, limit int) ([]domain.LocationHistoryEntry, error)
	ListNearby(ctx context.Context, tenantID uuid.UUID, lat, lon, radiusM float64, limit int) ([]domain.CourierLocation, error)
}

// TimelineRepo persists delivery timeline projections and geofence events.
type TimelineRepo interface {
	Append(ctx context.Context, e domain.TimelineEvent) error
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID, limit int) ([]domain.TimelineEvent, error)
	SaveGeofenceEvent(ctx context.Context, e domain.GeofenceEvent) error
	ListGeofenceEvents(ctx context.Context, tenantID uuid.UUID, courierID *uuid.UUID, limit int) ([]domain.GeofenceEvent, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// GeofenceCheckRequest asks geofence-service whether a point is inside zones.
type GeofenceCheckRequest struct {
	TenantID  uuid.UUID
	CourierID uuid.UUID
	OrderID   *uuid.UUID
	Lat       float64
	Lon       float64
}

// GeofenceHit is one zone enter/exit detection.
type GeofenceHit struct {
	ZoneID string
	Kind   domain.GeofenceEventKind
}

// GeofenceCheckResult is the response from GeofenceClient.
type GeofenceCheckResult struct {
	Hits []GeofenceHit
}

// GeofenceClient checks zone membership via geofence-service.
type GeofenceClient interface {
	Check(ctx context.Context, req GeofenceCheckRequest) (GeofenceCheckResult, error)
}
