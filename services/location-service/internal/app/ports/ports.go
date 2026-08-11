// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
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

// AddressRepo persists normalized addresses.
type AddressRepo interface {
	Upsert(ctx context.Context, a domain.NormalizedAddress) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.NormalizedAddress, error)
}

// POIRepo persists spatial POI index entries.
type POIRepo interface {
	Upsert(ctx context.Context, p domain.POI) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.POI, error)
	Nearby(ctx context.Context, q domain.NearbyQuery) ([]domain.POI, error)
	NearestOfKind(ctx context.Context, tenantID uuid.UUID, kind domain.POIKind, lat, lng float64, limit int) ([]domain.POI, error)
	InBBox(ctx context.Context, tenantID uuid.UUID, bbox domain.BBox, kind *domain.POIKind, limit int) ([]domain.POI, error)
	Count(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// HistoryRepo persists capped location history.
type HistoryRepo interface {
	Ingest(ctx context.Context, h domain.LocationHistory) error
	List(ctx context.Context, tenantID uuid.UUID, subjectType domain.SubjectType, subjectID string, limit int) ([]domain.LocationHistory, error)
}

// CacheRepo stores geocode cache and offline manifests.
type CacheRepo interface {
	GetGeocode(ctx context.Context, queryHash string) (domain.GeocodeResult, bool, error)
	SetGeocode(ctx context.Context, queryHash string, result domain.GeocodeResult, expiresAt time.Time) error
	GetOfflineManifest(ctx context.Context, tenantID uuid.UUID, region string) (domain.OfflineManifest, error)
	UpsertOfflineManifest(ctx context.Context, m domain.OfflineManifest) error
}

// HeatRepo persists demand heat cells.
type HeatRepo interface {
	Upsert(ctx context.Context, c domain.HeatCell) error
	List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.HeatCell, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// MapsProvider is the map SDK provider facade (no tile serving).
type MapsProvider interface {
	Geocode(ctx context.Context, query string) (domain.GeocodeResult, error)
	Reverse(ctx context.Context, lat, lng float64) (domain.GeocodeResult, error)
	Autocomplete(ctx context.Context, query string, limit int) ([]domain.GeocodeResult, error)
}

// GeofenceClient delegates zone serviceability to geofence-service.
type GeofenceClient interface {
	CheckServiceability(ctx context.Context, tenantID uuid.UUID, lat, lng float64) (domain.DeliveryFeasibility, error)
	Contains(ctx context.Context, tenantID, zoneID uuid.UUID, lat, lng float64) (bool, error)
}

// CreateRouteRequest proxies a route creation to routing-service.
type CreateRouteRequest struct {
	TenantID  uuid.UUID
	Origin    domain.LatLng
	Dest      domain.LatLng
	Waypoints []domain.LatLng
}

// CreateRouteResult is a routing-service route proxy response.
type CreateRouteResult struct {
	RouteID         string
	DistanceMeters  float64
	DurationSeconds float64
	Provider        string
}

// ETARequest proxies an ETA query to routing-service.
type ETARequest struct {
	TenantID uuid.UUID
	Origin   domain.LatLng
	Dest     domain.LatLng
}

// ETAResult is a routing-service ETA proxy response.
type ETAResult struct {
	DistanceMeters  float64
	DurationSeconds float64
	Provider        string
}

// RoutingClient delegates route/ETA to routing-service.
type RoutingClient interface {
	CreateRoute(ctx context.Context, req CreateRouteRequest) (CreateRouteResult, error)
	ETA(ctx context.Context, req ETARequest) (ETAResult, error)
}

// GeoSearchIndexer dual-writes POI/address documents to OpenSearch (geo_point).
type GeoSearchIndexer interface {
	IndexPOI(ctx context.Context, p domain.POI) error
	DeletePOI(ctx context.Context, tenantID, poiID uuid.UUID) error
	IndexAddress(ctx context.Context, a domain.NormalizedAddress) error
	SearchPOI(ctx context.Context, tenantID uuid.UUID, query string, lat, lng, radiusM float64, limit int) ([]uuid.UUID, error)
}
