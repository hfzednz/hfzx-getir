package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
)

// SearchStock queries the stock search index.
func (d *Deps) SearchStock(ctx context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	if d.Search == nil {
		return ports.SearchResult{}, nil
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	return d.Search.Search(ctx, q)
}

// ReindexStock reindexes all stock for a tenant.
func (d *Deps) ReindexStock(ctx context.Context, tenantID uuid.UUID) error {
	if d.Search == nil {
		return nil
	}
	return d.Search.ReindexAll(ctx, tenantID)
}
