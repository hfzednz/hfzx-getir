// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/domain"
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

// RouteRepo persists routes, legs, ETA snapshots, and traffic hints.
type RouteRepo interface {
	SaveRoute(ctx context.Context, r domain.Route) error
	GetRoute(ctx context.Context, tenantID, routeID uuid.UUID) (domain.Route, error)
	ListRoutes(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.Route, error)

	SaveETASnapshot(ctx context.Context, s domain.ETASnapshot) error
	ListETASnapshots(ctx context.Context, tenantID, routeID uuid.UUID, limit int) ([]domain.ETASnapshot, error)

	UpsertTrafficHint(ctx context.Context, h domain.TrafficHint) error
	GetTrafficHint(ctx context.Context, tenantID, id uuid.UUID) (domain.TrafficHint, error)
	ListActiveTrafficHints(ctx context.Context, tenantID uuid.UUID, at time.Time) ([]domain.TrafficHint, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// LatLon is a geographic point.
type LatLon struct {
	Lat float64
	Lon float64
}

// DistanceMatrixRequest asks for pairwise distances between origins and destinations.
type DistanceMatrixRequest struct {
	Origins      []LatLon
	Destinations []LatLon
}

// DistanceMatrixResult holds a distance (meters) and duration (seconds) matrix.
type DistanceMatrixResult struct {
	DistancesMeters  [][]float64
	DurationsSeconds [][]float64
}

// MapsClient provides distance/duration matrices (Google Maps stub OK).
type MapsClient interface {
	DistanceMatrix(ctx context.Context, req DistanceMatrixRequest) (DistanceMatrixResult, error)
}

// TrafficFactorRequest asks for a traffic multiplier near a point.
type TrafficFactorRequest struct {
	TenantID uuid.UUID
	Lat      float64
	Lon      float64
	At       time.Time
}

// TrafficClient returns traffic factors (stub OK).
type TrafficClient interface {
	Factor(ctx context.Context, req TrafficFactorRequest) (float64, error)
}

// WeatherFactorRequest asks for a weather multiplier near a point.
type WeatherFactorRequest struct {
	TenantID uuid.UUID
	Lat      float64
	Lon      float64
	At       time.Time
}

// WeatherClient returns weather factors (stub OK).
type WeatherClient interface {
	Factor(ctx context.Context, req WeatherFactorRequest) (float64, error)
}
