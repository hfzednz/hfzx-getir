package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// SearchProducts queries the search index.
func (d *Deps) SearchProducts(ctx context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	if d.Search == nil {
		return ports.SearchResult{}, domain.ErrNotFound
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	return d.Search.Search(ctx, q)
}

// SuggestProducts returns autocomplete suggestions.
func (d *Deps) SuggestProducts(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]string, error) {
	if d.Search == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	return d.Search.Suggest(ctx, tenantID, prefix, limit)
}

// ReindexTenant triggers a full tenant reindex (stub walks products in memory/pg).
func (d *Deps) ReindexTenant(ctx context.Context, tenantID uuid.UUID) error {
	if d.Search == nil {
		return nil
	}
	if err := d.Search.ReindexAll(ctx, tenantID); err != nil {
		return err
	}
	products, _, err := d.Products.List(ctx, ports.ProductFilter{TenantID: tenantID, Limit: 10000})
	if err != nil {
		return err
	}
	for _, p := range products {
		doc := d.buildSearchDoc(ctx, p)
		_ = d.Search.IndexProduct(ctx, doc)
	}
	return nil
}
