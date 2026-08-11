package traffic

import (
	"context"
	"log/slog"

	"github.com/nexora/routing-service/internal/app/ports"
)

// Client is a traffic factor stub (default 1.0).
type Client struct {
	factor float64
	log    *slog.Logger
}

// NewClient returns a fixed-factor TrafficClient.
func NewClient(log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{factor: 1.0, log: log}
}

// Factor returns the configured traffic multiplier.
func (c *Client) Factor(_ context.Context, req ports.TrafficFactorRequest) (float64, error) {
	c.log.Debug("traffic.factor", "lat", req.Lat, "lon", req.Lon, "factor", c.factor)
	return c.factor, nil
}

var _ ports.TrafficClient = (*Client)(nil)
