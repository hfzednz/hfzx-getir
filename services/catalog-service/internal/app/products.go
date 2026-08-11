package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateProductInput creates a draft master product.
type CreateProductInput struct {
	TenantID    uuid.UUID
	BrandID     *uuid.UUID
	Kind        domain.ProductKind
	Slug        string
	SKUCode     string
	ExternalRef string
	Metadata    map[string]any
}

// CreateProduct creates a new draft product.
func (d *Deps) CreateProduct(ctx context.Context, in CreateProductInput) (domain.Product, error) {
	now := d.now()
	if in.Kind == "" {
		in.Kind = domain.ProductKindStandard
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	p := domain.Product{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		BrandID:     in.BrandID,
		Kind:        in.Kind,
		Status:      domain.ProductStatusDraft,
		Slug:        strings.TrimSpace(in.Slug),
		SKUCode:     strings.TrimSpace(in.SKUCode),
		ExternalRef: strings.TrimSpace(in.ExternalRef),
		Metadata:    in.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.Validate(); err != nil {
		return domain.Product{}, err
	}
	if _, err := d.Products.GetBySlug(ctx, in.TenantID, p.Slug); err == nil {
		return domain.Product{}, domain.ErrAlreadyExists
	}
	if err := d.Products.Create(ctx, p); err != nil {
		return domain.Product{}, err
	}
	d.publishEvent(ctx, domain.EventProductCreated, p.TenantID, p.ID, map[string]any{"productId": p.ID, "status": p.Status})
	return p, nil
}

// UpdateProductInput patches an editable product.
type UpdateProductInput struct {
	TenantID        uuid.UUID
	ProductID       uuid.UUID
	BrandID         *uuid.UUID
	Slug            *string
	SKUCode         *string
	ExternalRef     *string
	GTINBase        *string
	ManufacturerSKU *string
	Metadata        map[string]any
}

// UpdateProduct updates product fields when editable.
func (d *Deps) UpdateProduct(ctx context.Context, in UpdateProductInput) (domain.Product, error) {
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.Product{}, err
	}
	if err := d.ensureEditable(p); err != nil {
		return domain.Product{}, err
	}
	if in.BrandID != nil {
		p.BrandID = in.BrandID
	}
	if in.Slug != nil {
		p.Slug = strings.TrimSpace(*in.Slug)
	}
	if in.SKUCode != nil {
		p.SKUCode = strings.TrimSpace(*in.SKUCode)
	}
	if in.ExternalRef != nil {
		p.ExternalRef = strings.TrimSpace(*in.ExternalRef)
	}
	if in.GTINBase != nil {
		p.GTINBase = strings.TrimSpace(*in.GTINBase)
	}
	if in.ManufacturerSKU != nil {
		p.ManufacturerSKU = strings.TrimSpace(*in.ManufacturerSKU)
	}
	if in.Metadata != nil {
		p.Metadata = in.Metadata
	}
	p.UpdatedAt = d.now()
	if err := p.Validate(); err != nil {
		return domain.Product{}, err
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return domain.Product{}, err
	}
	d.publishEvent(ctx, domain.EventProductUpdated, p.TenantID, p.ID, map[string]any{"productId": p.ID})
	d.indexProduct(ctx, p.TenantID, p.ID)
	return p, nil
}

// GetProduct returns a product by id.
func (d *Deps) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.Product, error) {
	return d.getProduct(ctx, tenantID, productID)
}

// ListProducts lists products with filters.
func (d *Deps) ListProducts(ctx context.Context, f ports.ProductFilter) ([]domain.Product, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	return d.Products.List(ctx, f)
}

// ArchiveProduct moves a product to archived status.
func (d *Deps) ArchiveProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.Product, error) {
	return d.applyStatus(ctx, tenantID, productID, domain.ProductStatusArchived, domain.EventProductArchived)
}

func (d *Deps) applyStatus(ctx context.Context, tenantID, productID uuid.UUID, to domain.ProductStatus, eventType string) (domain.Product, error) {
	p, err := d.getProduct(ctx, tenantID, productID)
	if err != nil {
		return domain.Product{}, err
	}
	now := d.now()
	if err := p.TransitionTo(to, now); err != nil {
		return domain.Product{}, err
	}
	if err := p.Validate(); err != nil {
		return domain.Product{}, err
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return domain.Product{}, err
	}
	if eventType != "" {
		d.publishEvent(ctx, eventType, tenantID, productID, map[string]any{"status": to})
	}
	d.indexProduct(ctx, tenantID, productID)
	return p, nil
}

// DeleteProduct soft-deletes a product.
func (d *Deps) DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error {
	p, err := d.getProduct(ctx, tenantID, productID)
	if err != nil {
		return err
	}
	now := d.now()
	if err := p.TransitionTo(domain.ProductStatusDeleted, now); err != nil {
		return err
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return err
	}
	if d.Search != nil {
		_ = d.Search.DeleteProduct(ctx, tenantID, productID)
	}
	return nil
}

// TransitionProductStatus applies an arbitrary valid status transition (admin).
func (d *Deps) TransitionProductStatus(ctx context.Context, tenantID, productID uuid.UUID, to domain.ProductStatus) (domain.Product, error) {
	p, err := d.getProduct(ctx, tenantID, productID)
	if err != nil {
		return domain.Product{}, err
	}
	now := d.now()
	if err := p.TransitionTo(to, now); err != nil {
		return domain.Product{}, err
	}
	if err := p.Validate(); err != nil {
		return domain.Product{}, fmt.Errorf("%w: %v", domain.ErrInvariant, err)
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return domain.Product{}, err
	}
	d.indexProduct(ctx, tenantID, productID)
	return p, nil
}
