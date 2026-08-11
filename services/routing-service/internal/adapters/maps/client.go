package maps

import (
	"context"
	"log/slog"

	"github.com/nexora/routing-service/internal/app"
	"github.com/nexora/routing-service/internal/app/ports"
)

// Client is a Google Maps distance-matrix stub (Haversine-backed).
type Client struct {
	inner app.HaversineMapsClient
	log   *slog.Logger
}

// NewClient returns a Haversine-backed MapsClient.
func NewClient(log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{inner: app.HaversineMapsClient{}, log: log}
}

// DistanceMatrix returns pairwise Haversine distances/durations.
func (c *Client) DistanceMatrix(ctx context.Context, req ports.DistanceMatrixRequest) (ports.DistanceMatrixResult, error) {
	c.log.Debug("maps.distance_matrix", "origins", len(req.Origins), "destinations", len(req.Destinations))
	return c.inner.DistanceMatrix(ctx, req)
}

var _ ports.MapsClient = (*Client)(nil)
