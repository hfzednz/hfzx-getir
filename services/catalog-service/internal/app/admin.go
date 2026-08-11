package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// BulkUpdateStatusInput bulk-updates product status for explorer/admin.
type BulkUpdateStatusInput struct {
	TenantID   uuid.UUID
	ProductIDs []uuid.UUID
	ToStatus   domain.ProductStatus
	ActorID    uuid.UUID
}

// BulkUpdateStatus applies a status transition to many products.
func (d *Deps) BulkUpdateStatus(ctx context.Context, in BulkUpdateStatusInput) ([]domain.Product, []error) {
	results := make([]domain.Product, 0, len(in.ProductIDs))
	errs := make([]error, 0)
	for _, id := range in.ProductIDs {
		p, err := d.TransitionProductStatus(ctx, in.TenantID, id, in.ToStatus)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, p)
	}
	return results, errs
}

// DuplicateCandidate groups potential duplicate products.
type DuplicateCandidate struct {
	Key      string
	Products []domain.Product
}

// FindDuplicates scans for slug/sku collisions within a tenant.
func (d *Deps) FindDuplicates(ctx context.Context, tenantID uuid.UUID) ([]DuplicateCandidate, error) {
	products, _, err := d.Products.List(ctx, ports.ProductFilter{TenantID: tenantID, Limit: 10000})
	if err != nil {
		return nil, err
	}
	bySlug := map[string][]domain.Product{}
	bySKU := map[string][]domain.Product{}
	for _, p := range products {
		if p.Status == domain.ProductStatusDeleted {
			continue
		}
		bySlug[p.Slug] = append(bySlug[p.Slug], p)
		if p.SKUCode != "" {
			bySKU[p.SKUCode] = append(bySKU[p.SKUCode], p)
		}
	}
	out := make([]DuplicateCandidate, 0)
	for slug, group := range bySlug {
		if len(group) > 1 {
			out = append(out, DuplicateCandidate{Key: "slug:" + slug, Products: group})
		}
	}
	for sku, group := range bySKU {
		if len(group) > 1 {
			out = append(out, DuplicateCandidate{Key: "sku:" + sku, Products: group})
		}
	}
	return out, nil
}

// ExplorerFilter is admin product explorer query.
type ExplorerFilter struct {
	TenantID uuid.UUID
	Query    string
	Status   *domain.ProductStatus
	Limit    int
	Offset   int
}

// ExplorerSearch lists products for admin explorer with text filter.
func (d *Deps) ExplorerSearch(ctx context.Context, f ExplorerFilter) ([]domain.Product, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	f.Query = strings.TrimSpace(f.Query)
	return d.Products.List(ctx, ports.ProductFilter{
		TenantID: f.TenantID,
		Query:    f.Query,
		Status:   f.Status,
		Limit:    f.Limit,
		Offset:   f.Offset,
	})
}
