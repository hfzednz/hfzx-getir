package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/domain"
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

// DispatchRepo stores dispatches, events, attempts, and batches.
type DispatchRepo interface {
	Create(ctx context.Context, d domain.Dispatch) error
	Update(ctx context.Context, d domain.Dispatch) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error)
	List(ctx context.Context, tenantID uuid.UUID, status domain.DispatchStatus, limit, offset int) ([]domain.Dispatch, int, error)
	AppendEvent(ctx context.Context, e domain.DispatchEvent) error
	ListEvents(ctx context.Context, tenantID, dispatchID uuid.UUID) ([]domain.DispatchEvent, error)
	AppendAttempt(ctx context.Context, a domain.AssignmentAttempt) error
	CreateBatch(ctx context.Context, b domain.Batch) error
	GetBatch(ctx context.Context, tenantID, id uuid.UUID) (domain.Batch, error)
}

// CourierPool manages courier availability snapshots.
type CourierPool interface {
	Upsert(ctx context.Context, c domain.CourierSnapshot) error
	Get(ctx context.Context, tenantID, courierPrincipalID uuid.UUID) (domain.CourierSnapshot, error)
	ListAvailable(ctx context.Context, tenantID uuid.UUID) ([]domain.CourierSnapshot, error)
	AdjustLoad(ctx context.Context, tenantID, courierPrincipalID uuid.UUID, delta int) error
}

// VehicleRepo stores fleet vehicles.
type VehicleRepo interface {
	Upsert(ctx context.Context, v domain.Vehicle) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Vehicle, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Vehicle, int, error)
}

// RouteRequest is input for routing-service CreateRoute.
type RouteRequest struct {
	TenantID    uuid.UUID
	DispatchID  uuid.UUID
	Waypoints   []domain.Point
}

// RouteResult is a stub route plan.
type RouteResult struct {
	RouteID    uuid.UUID
	ETASeconds int
	DistanceM  float64
}

// RoutingClient talks to routing-service (CreateRoute / ETA).
type RoutingClient interface {
	CreateRoute(ctx context.Context, req RouteRequest) (RouteResult, error)
	EstimateETA(ctx context.Context, from, to domain.Point) (int, error)
}

// TrackingClient talks to tracking-service (subscribe stub).
type TrackingClient interface {
	SubscribeDispatch(ctx context.Context, tenantID, dispatchID, courierPrincipalID uuid.UUID) error
}

// GeofenceClient talks to geofence-service.
type GeofenceClient interface {
	CheckServiceability(ctx context.Context, tenantID uuid.UUID, city string, p domain.Point) (bool, error)
}
