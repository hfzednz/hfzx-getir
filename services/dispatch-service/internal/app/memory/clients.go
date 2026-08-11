package memory

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// RoutingClient is a stub routing client.
type RoutingClient struct {
	RouteID uuid.UUID
	ETA     int
}

func (c *RoutingClient) CreateRoute(_ context.Context, req ports.RouteRequest) (ports.RouteResult, error) {
	id := c.RouteID
	if id == uuid.Nil {
		id = uuid.New()
	}
	eta := c.ETA
	if eta == 0 {
		eta = 900
	}
	dist := 0.0
	if len(req.Waypoints) >= 2 {
		dist = domain.HaversineMeters(req.Waypoints[0], req.Waypoints[1])
	}
	return ports.RouteResult{RouteID: id, ETASeconds: eta, DistanceM: dist}, nil
}

func (c *RoutingClient) EstimateETA(_ context.Context, from, to domain.Point) (int, error) {
	_ = from
	_ = to
	if c.ETA > 0 {
		return c.ETA, nil
	}
	return 600, nil
}

// TrackingClient is a subscribe stub.
type TrackingClient struct {
	Subscribed []uuid.UUID
}

func (c *TrackingClient) SubscribeDispatch(_ context.Context, _, dispatchID, _ uuid.UUID) error {
	c.Subscribed = append(c.Subscribed, dispatchID)
	return nil
}

// GeofenceClient is a serviceability stub.
type GeofenceClient struct {
	OK bool
}

func (c *GeofenceClient) CheckServiceability(_ context.Context, _ uuid.UUID, _ string, _ domain.Point) (bool, error) {
	return c.OK, nil
}

var (
	_ ports.RoutingClient  = (*RoutingClient)(nil)
	_ ports.TrackingClient = (*TrackingClient)(nil)
	_ ports.GeofenceClient = (*GeofenceClient)(nil)
)
