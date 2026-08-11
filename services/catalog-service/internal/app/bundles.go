package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// UpsertBundleInput defines bundle composition for a bundle/kit product.
type UpsertBundleInput struct {
	TenantID    uuid.UUID
	ProductID   uuid.UUID
	Composition domain.BundleCompositionType
	Name        string
	Items       []BundleItemInput
}

// BundleItemInput is a bundle component line.
type BundleItemInput struct {
	ComponentVariantID uuid.UUID
	Qty                int
	IsOptional         bool
	SortOrder          int
}

// UpsertBundle creates or updates bundle composition.
func (d *Deps) UpsertBundle(ctx context.Context, in UpsertBundleInput) (domain.Bundle, []domain.BundleItem, error) {
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.Bundle{}, nil, err
	}
	if err := d.ensureEditable(p); err != nil {
		return domain.Bundle{}, nil, err
	}
	if in.Composition == "" {
		in.Composition = domain.BundleCompositionStatic
	}
	now := d.now()
	b := domain.Bundle{
		ID:          d.newID(),
		ProductID:   in.ProductID,
		TenantID:    in.TenantID,
		Composition: in.Composition,
		Name:        strings.TrimSpace(in.Name),
		Metadata:    map[string]any{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if existing, err := d.Bundles.GetByProduct(ctx, in.TenantID, in.ProductID); err == nil {
		b.ID = existing.ID
		b.CreatedAt = existing.CreatedAt
	}
	if err := b.Validate(); err != nil {
		return domain.Bundle{}, nil, err
	}
	if err := d.Bundles.Upsert(ctx, b); err != nil {
		return domain.Bundle{}, nil, err
	}
	items := make([]domain.BundleItem, 0, len(in.Items))
	for _, it := range in.Items {
		item := domain.BundleItem{
			ID:                 d.newID(),
			BundleID:           b.ID,
			ComponentVariantID: it.ComponentVariantID,
			Qty:                it.Qty,
			IsOptional:         it.IsOptional,
			SortOrder:          it.SortOrder,
			Metadata:           map[string]any{},
			CreatedAt:          now,
		}
		if err := item.Validate(); err != nil {
			return domain.Bundle{}, nil, err
		}
		items = append(items, item)
	}
	if err := d.Bundles.SetItems(ctx, b.ID, items); err != nil {
		return domain.Bundle{}, nil, err
	}
	event := domain.EventBundleCreated
	if b.UpdatedAt.After(b.CreatedAt) {
		event = domain.EventBundleUpdated
	}
	d.publishEvent(ctx, event, in.TenantID, in.ProductID, map[string]any{"bundleId": b.ID})
	return b, items, nil
}

// GetBundle returns bundle header and items.
func (d *Deps) GetBundle(ctx context.Context, tenantID, productID uuid.UUID) (domain.Bundle, []domain.BundleItem, error) {
	b, err := d.Bundles.GetByProduct(ctx, tenantID, productID)
	if err != nil {
		return domain.Bundle{}, nil, err
	}
	items, err := d.Bundles.ListItems(ctx, b.ID)
	if err != nil {
		return domain.Bundle{}, nil, err
	}
	return b, items, nil
}
