package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// MockMapsProvider is a deterministic MapsProvider for tests and memory mode.
type MockMapsProvider struct {
	GeocodeCalls      int
	ReverseCalls      int
	AutocompleteCalls int
}

func (m *MockMapsProvider) Geocode(_ context.Context, query string) (domain.GeocodeResult, error) {
	m.GeocodeCalls++
	q := strings.TrimSpace(query)
	if q == "" {
		return domain.GeocodeResult{}, fmt.Errorf("%w: query required", domain.ErrInvalidArgument)
	}
	return domain.GeocodeResult{
		PlaceID:    "mock:place:" + hashShort(q),
		Formatted:  q,
		Lat:        41.0082,
		Lng:        28.9784,
		Confidence: 0.92,
		Components: domain.AddressComponents{Country: "TR", City: "Istanbul", Street: q},
		Provider:   "mock",
	}, nil
}

func (m *MockMapsProvider) Reverse(_ context.Context, lat, lng float64) (domain.GeocodeResult, error) {
	m.ReverseCalls++
	if !domain.ValidLatLng(lat, lng) {
		return domain.GeocodeResult{}, fmt.Errorf("%w: lat/lng out of range", domain.ErrInvalidArgument)
	}
	return domain.GeocodeResult{
		PlaceID:    fmt.Sprintf("mock:rev:%.4f:%.4f", lat, lng),
		Formatted:  "Mock Street 1, Istanbul",
		Lat:        lat,
		Lng:        lng,
		Confidence: 0.88,
		Components: domain.AddressComponents{Country: "TR", City: "Istanbul", Street: "Mock Street"},
		Provider:   "mock",
	}, nil
}

func (m *MockMapsProvider) Autocomplete(_ context.Context, query string, limit int) ([]domain.GeocodeResult, error) {
	m.AutocompleteCalls++
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("%w: query required", domain.ErrInvalidArgument)
	}
	if limit <= 0 {
		limit = 5
	}
	out := make([]domain.GeocodeResult, 0, limit)
	for i := 0; i < limit && i < 3; i++ {
		out = append(out, domain.GeocodeResult{
			PlaceID:    fmt.Sprintf("mock:ac:%s:%d", hashShort(q), i),
			Formatted:  fmt.Sprintf("%s Suggestion %d", q, i+1),
			Lat:        41.0082 + float64(i)*0.001,
			Lng:        28.9784 + float64(i)*0.001,
			Confidence: domain.ConfidenceScore(0.8 - float64(i)*0.05),
			Components: domain.AddressComponents{Country: "TR", City: "Istanbul"},
			Provider:   "mock",
		})
	}
	return out, nil
}

func hashShort(s string) string {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return fmt.Sprintf("%08x", h)
}

// MemoryGeofenceClient is an in-process GeofenceClient stub.
type MemoryGeofenceClient struct {
	Serviceable bool
	Calls       int
	LastLat     float64
	LastLng     float64
	LastTenant  uuid.UUID
	ZoneID      uuid.UUID
}

// NewMemoryGeofenceClient returns a serviceable-by-default geofence stub.
func NewMemoryGeofenceClient() *MemoryGeofenceClient {
	return &MemoryGeofenceClient{Serviceable: true}
}

func (c *MemoryGeofenceClient) CheckServiceability(_ context.Context, tenantID uuid.UUID, lat, lng float64) (domain.DeliveryFeasibility, error) {
	c.Calls++
	c.LastTenant = tenantID
	c.LastLat = lat
	c.LastLng = lng
	feas := domain.DeliveryFeasibility{
		Serviceable: c.Serviceable,
		Reason:      "memory-geofence",
		Score:       1,
	}
	if !c.Serviceable {
		feas.Score = 0
		feas.Reason = "out_of_zone"
	}
	if feas.Serviceable && c.ZoneID != uuid.Nil {
		zid := c.ZoneID
		feas.ZoneID = &zid
	}
	return feas, nil
}

func (c *MemoryGeofenceClient) Contains(_ context.Context, _, _ uuid.UUID, _, _ float64) (bool, error) {
	return c.Serviceable, nil
}

// MemoryRoutingClient records CreateRoute/ETA calls for tests.
type MemoryRoutingClient struct {
	CreateRouteCalls int
	ETACalls         int
	LastCreate       ports.CreateRouteRequest
	LastETA          ports.ETARequest
}

func (c *MemoryRoutingClient) CreateRoute(_ context.Context, req ports.CreateRouteRequest) (ports.CreateRouteResult, error) {
	c.CreateRouteCalls++
	c.LastCreate = req
	dist := domain.HaversineDistanceMeters(req.Origin, req.Dest)
	return ports.CreateRouteResult{
		RouteID:         uuid.NewString(),
		DistanceMeters:  dist,
		DurationSeconds: dist / 8.33,
		Provider:        "memory-routing",
	}, nil
}

func (c *MemoryRoutingClient) ETA(_ context.Context, req ports.ETARequest) (ports.ETAResult, error) {
	c.ETACalls++
	c.LastETA = req
	dist := domain.HaversineDistanceMeters(req.Origin, req.Dest)
	return ports.ETAResult{
		DistanceMeters:  dist,
		DurationSeconds: dist / 8.33,
		Provider:        "memory-routing",
	}, nil
}

var (
	_ ports.MapsProvider   = (*MockMapsProvider)(nil)
	_ ports.GeofenceClient = (*MemoryGeofenceClient)(nil)
	_ ports.RoutingClient  = (*MemoryRoutingClient)(nil)
)
