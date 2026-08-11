package media

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/memory"
)

// Client validates media refs via media-service (mock in dev).
type Client struct {
	BaseURL string
	inner   memory.MockMedia
}

// NewClient returns a media port client.
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// ValidateRef checks a media asset reference.
func (c *Client) ValidateRef(ctx context.Context, tenantID uuid.UUID, mediaRef string) (bool, string, error) {
	return c.inner.ValidateRef(ctx, tenantID, mediaRef)
}
