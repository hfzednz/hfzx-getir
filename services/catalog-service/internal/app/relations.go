package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// SetProductRelationInput creates or updates a product relation.
type SetProductRelationInput struct {
	TenantID        uuid.UUID
	SourceProductID uuid.UUID
	TargetProductID uuid.UUID
	Type            domain.RelationType
	SortOrder       int
	Score           *float64
}

// SetProductRelation upserts a merchandising relation.
func (d *Deps) SetProductRelation(ctx context.Context, in SetProductRelationInput) (domain.ProductRelation, error) {
	if _, err := d.getProduct(ctx, in.TenantID, in.SourceProductID); err != nil {
		return domain.ProductRelation{}, err
	}
	if _, err := d.getProduct(ctx, in.TenantID, in.TargetProductID); err != nil {
		return domain.ProductRelation{}, err
	}
	now := d.now()
	r := domain.ProductRelation{
		ID:              d.newID(),
		TenantID:        in.TenantID,
		SourceProductID: in.SourceProductID,
		TargetProductID: in.TargetProductID,
		Type:            in.Type,
		SortOrder:       in.SortOrder,
		Score:           in.Score,
		Metadata:        map[string]any{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := r.Validate(); err != nil {
		return domain.ProductRelation{}, err
	}
	if err := d.Relations.Upsert(ctx, r); err != nil {
		return domain.ProductRelation{}, err
	}
	return r, nil
}

// ListProductRelations lists outgoing relations from a product.
func (d *Deps) ListProductRelations(ctx context.Context, tenantID, sourceProductID uuid.UUID, typ *domain.RelationType) ([]domain.ProductRelation, error) {
	return d.Relations.ListBySource(ctx, tenantID, sourceProductID, typ)
}

// DeleteProductRelation removes a relation.
func (d *Deps) DeleteProductRelation(ctx context.Context, tenantID, relationID uuid.UUID) error {
	return d.Relations.Delete(ctx, tenantID, relationID)
}
