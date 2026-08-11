package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateAttributeDefInput defines a tenant attribute schema.
type CreateAttributeDefInput struct {
	TenantID      uuid.UUID
	Code          string
	Name          string
	Description   string
	Type          domain.AttributeType
	Schema        map[string]any
	IsRequired    bool
	IsFilterable  bool
	IsVariantAxis bool
	SortOrder     int
}

// CreateAttributeDef creates an attribute definition.
func (d *Deps) CreateAttributeDef(ctx context.Context, in CreateAttributeDefInput) (domain.AttributeDef, error) {
	if in.Type == "" {
		in.Type = domain.AttributeTypeText
	}
	if in.Schema == nil {
		in.Schema = map[string]any{}
	}
	now := d.now()
	def := domain.AttributeDef{
		ID:            d.newID(),
		TenantID:      in.TenantID,
		Code:          strings.TrimSpace(in.Code),
		Name:          strings.TrimSpace(in.Name),
		Description:   strings.TrimSpace(in.Description),
		Type:          in.Type,
		Schema:        in.Schema,
		IsRequired:    in.IsRequired,
		IsFilterable:  in.IsFilterable,
		IsVariantAxis: in.IsVariantAxis,
		SortOrder:     in.SortOrder,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := def.Validate(); err != nil {
		return domain.AttributeDef{}, err
	}
	if _, err := d.Attributes.GetDefByCode(ctx, in.TenantID, def.Code); err == nil {
		return domain.AttributeDef{}, domain.ErrAlreadyExists
	}
	if err := d.Attributes.CreateDef(ctx, def); err != nil {
		return domain.AttributeDef{}, err
	}
	return def, nil
}

// ListAttributeDefs lists attribute definitions.
func (d *Deps) ListAttributeDefs(ctx context.Context, tenantID uuid.UUID) ([]domain.AttributeDef, error) {
	return d.Attributes.ListDefs(ctx, tenantID)
}

// SetProductAttributeInput sets a product attribute value.
type SetProductAttributeInput struct {
	TenantID       uuid.UUID
	ProductID      uuid.UUID
	AttributeDefID uuid.UUID
	Value          map[string]any
	Locale         string
}

// SetProductAttribute upserts a product attribute value.
func (d *Deps) SetProductAttribute(ctx context.Context, in SetProductAttributeInput) (domain.ProductAttribute, error) {
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.ProductAttribute{}, err
	}
	if err := d.ensureEditable(p); err != nil {
		return domain.ProductAttribute{}, err
	}
	if in.Value == nil {
		in.Value = map[string]any{}
	}
	now := d.now()
	a := domain.ProductAttribute{
		ID:             d.newID(),
		ProductID:      in.ProductID,
		AttributeDefID: in.AttributeDefID,
		TenantID:       in.TenantID,
		Value:          in.Value,
		Locale:         strings.TrimSpace(in.Locale),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := a.Validate(); err != nil {
		return domain.ProductAttribute{}, err
	}
	if err := d.Attributes.UpsertProductAttribute(ctx, a); err != nil {
		return domain.ProductAttribute{}, err
	}
	d.indexProduct(ctx, in.TenantID, in.ProductID)
	return a, nil
}

// ListProductAttributes lists attribute values for a product.
func (d *Deps) ListProductAttributes(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductAttribute, error) {
	return d.Attributes.ListProductAttributes(ctx, tenantID, productID)
}
