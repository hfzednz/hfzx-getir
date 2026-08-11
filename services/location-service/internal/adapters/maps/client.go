package maps

import (
	"context"
	"log/slog"

	"github.com/nexora/location-service/internal/app"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// Client is a maps provider facade stub (no tile serving).
type Client struct {
	inner *app.MockMapsProvider
	log   *slog.Logger
}

// NewClient returns a MockMapsProvider-backed MapsProvider.
func NewClient(log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{inner: &app.MockMapsProvider{}, log: log}
}

func (c *Client) Geocode(ctx context.Context, query string) (domain.GeocodeResult, error) {
	c.log.Debug("maps.geocode", "queryLen", len(query))
	return c.inner.Geocode(ctx, query)
}

func (c *Client) Reverse(ctx context.Context, lat, lng float64) (domain.GeocodeResult, error) {
	c.log.Debug("maps.reverse", "note", "coords omitted for privacy")
	return c.inner.Reverse(ctx, lat, lng)
}

func (c *Client) Autocomplete(ctx context.Context, query string, limit int) ([]domain.GeocodeResult, error) {
	c.log.Debug("maps.autocomplete", "queryLen", len(query), "limit", limit)
	return c.inner.Autocomplete(ctx, query, limit)
}

var _ ports.MapsProvider = (*Client)(nil)
