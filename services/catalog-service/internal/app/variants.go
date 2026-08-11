package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateVariantInput creates a variant under a product.
type CreateVariantInput struct {
	TenantID     uuid.UUID
	ProductID    uuid.UUID
	SKUCode      string
	Name         string
	OptionValues map[string]any
	SortOrder    int
}

// CreateVariant adds a variant to a product.
func (d *Deps) CreateVariant(ctx context.Context, in CreateVariantInput) (domain.Variant, error) {
	if _, err := d.getProduct(ctx, in.TenantID, in.ProductID); err != nil {
		return domain.Variant{}, err
	}
	if in.OptionValues == nil {
		in.OptionValues = map[string]any{}
	}
	now := d.now()
	v := domain.Variant{
		ID:           d.newID(),
		ProductID:    in.ProductID,
		TenantID:     in.TenantID,
		SKUCode:      strings.TrimSpace(in.SKUCode),
		Name:         strings.TrimSpace(in.Name),
		OptionValues: in.OptionValues,
		Status:       domain.VariantStatusDraft,
		SortOrder:    in.SortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := v.Validate(); err != nil {
		return domain.Variant{}, err
	}
	if err := d.Variants.Create(ctx, v); err != nil {
		return domain.Variant{}, err
	}
	d.publishEvent(ctx, domain.EventVariantCreated, in.TenantID, in.ProductID, map[string]any{"variantId": v.ID})
	return v, nil
}

// ListVariants returns variants for a product.
func (d *Deps) ListVariants(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.Variant, error) {
	if _, err := d.getProduct(ctx, tenantID, productID); err != nil {
		return nil, err
	}
	return d.Variants.ListByProduct(ctx, tenantID, productID)
}

// AddSKUIdentifierInput attaches a barcode/code to a variant.
type AddSKUIdentifierInput struct {
	TenantID  uuid.UUID
	VariantID uuid.UUID
	Type      domain.SKUIdentifierType
	Value     string
	IsPrimary bool
	Metadata  map[string]any
}

// AddSKUIdentifier creates a SKU identifier on a variant.
func (d *Deps) AddSKUIdentifier(ctx context.Context, in AddSKUIdentifierInput) (domain.SKUIdentifier, error) {
	v, err := d.Variants.GetByID(ctx, in.TenantID, in.VariantID)
	if err != nil {
		return domain.SKUIdentifier{}, err
	}
	value := domain.NormalizeBarcode(in.Type, in.Value)
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	now := d.now()
	s := domain.SKUIdentifier{
		ID:        d.newID(),
		VariantID: in.VariantID,
		TenantID:  in.TenantID,
		Type:      in.Type,
		Value:     value,
		IsPrimary: in.IsPrimary,
		Metadata:  in.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Validate(); err != nil {
		return domain.SKUIdentifier{}, err
	}
	if _, err := d.SKUs.FindByValue(ctx, in.TenantID, in.Type, value); err == nil {
		return domain.SKUIdentifier{}, domain.ErrAlreadyExists
	}
	if err := d.SKUs.Create(ctx, s); err != nil {
		return domain.SKUIdentifier{}, err
	}
	d.indexProduct(ctx, in.TenantID, v.ProductID)
	return s, nil
}

// FindByBarcode looks up a SKU identifier by type and value.
func (d *Deps) FindByBarcode(ctx context.Context, tenantID uuid.UUID, typ domain.SKUIdentifierType, value string) (domain.SKUIdentifier, domain.Variant, domain.Product, error) {
	value = domain.NormalizeBarcode(typ, value)
	s, err := d.SKUs.FindByValue(ctx, tenantID, typ, value)
	if err != nil {
		return domain.SKUIdentifier{}, domain.Variant{}, domain.Product{}, err
	}
	v, err := d.Variants.GetByID(ctx, tenantID, s.VariantID)
	if err != nil {
		return domain.SKUIdentifier{}, domain.Variant{}, domain.Product{}, err
	}
	p, err := d.getProduct(ctx, tenantID, v.ProductID)
	if err != nil {
		return domain.SKUIdentifier{}, domain.Variant{}, domain.Product{}, err
	}
	return s, v, p, nil
}

// ListSKUIdentifiers lists identifiers for a variant.
func (d *Deps) ListSKUIdentifiers(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.SKUIdentifier, error) {
	return d.SKUs.ListByVariant(ctx, tenantID, variantID)
}
