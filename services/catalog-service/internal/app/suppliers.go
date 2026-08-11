package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// CreateSupplierInput creates a supplier master.
type CreateSupplierInput struct {
	TenantID     uuid.UUID
	Code         string
	Name         string
	ContactEmail string
	ContactPhone string
	CountryCode  string
	ExternalRef  string
}

// CreateSupplier inserts a supplier.
func (d *Deps) CreateSupplier(ctx context.Context, in CreateSupplierInput) (domain.Supplier, error) {
	if d.Suppliers == nil {
		return domain.Supplier{}, domain.ErrInvalidArgument
	}
	now := d.now()
	s := domain.Supplier{
		ID:           d.newID(),
		TenantID:     in.TenantID,
		Code:         strings.TrimSpace(in.Code),
		Name:         strings.TrimSpace(in.Name),
		ContactEmail: strings.TrimSpace(in.ContactEmail),
		ContactPhone: strings.TrimSpace(in.ContactPhone),
		CountryCode:  strings.ToUpper(strings.TrimSpace(in.CountryCode)),
		ExternalRef:  strings.TrimSpace(in.ExternalRef),
		Metadata:     map[string]any{},
		Status:       domain.SupplierStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Validate(); err != nil {
		return domain.Supplier{}, err
	}
	if _, err := d.Suppliers.GetByCode(ctx, in.TenantID, s.Code); err == nil {
		return domain.Supplier{}, domain.ErrAlreadyExists
	}
	if err := d.Suppliers.Create(ctx, s); err != nil {
		return domain.Supplier{}, err
	}
	d.publishEvent(ctx, domain.EventSupplierChanged, in.TenantID, uuid.Nil, map[string]any{
		"supplierId": s.ID, "action": "created",
	})
	return s, nil
}

// GetSupplier returns a supplier by id.
func (d *Deps) GetSupplier(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	if d.Suppliers == nil {
		return domain.Supplier{}, domain.ErrNotFound
	}
	return d.Suppliers.GetByID(ctx, tenantID, id)
}

// ListSuppliers lists suppliers for a tenant.
func (d *Deps) ListSuppliers(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Supplier, int, error) {
	if d.Suppliers == nil {
		return nil, 0, nil
	}
	if limit <= 0 {
		limit = 50
	}
	return d.Suppliers.List(ctx, tenantID, limit, offset)
}

// LinkSupplierProductInput links a product to a supplier.
type LinkSupplierProductInput struct {
	TenantID         uuid.UUID
	SupplierID       uuid.UUID
	ProductID        uuid.UUID
	VariantID        *uuid.UUID
	SupplierSKU      string
	CostHintMinor    *int64
	CostHintCurrency string
	LeadTimeDays     *int
	MOQ              *int
	IsPreferred      bool
}

// LinkSupplierProduct attaches product metadata to a supplier.
func (d *Deps) LinkSupplierProduct(ctx context.Context, in LinkSupplierProductInput) (domain.SupplierProduct, error) {
	if d.Suppliers == nil {
		return domain.SupplierProduct{}, domain.ErrInvalidArgument
	}
	if _, err := d.Suppliers.GetByID(ctx, in.TenantID, in.SupplierID); err != nil {
		return domain.SupplierProduct{}, err
	}
	if _, err := d.getProduct(ctx, in.TenantID, in.ProductID); err != nil {
		return domain.SupplierProduct{}, err
	}
	now := d.now()
	sp := domain.SupplierProduct{
		ID:               d.newID(),
		SupplierID:       in.SupplierID,
		ProductID:        in.ProductID,
		VariantID:        in.VariantID,
		TenantID:         in.TenantID,
		SupplierSKU:      strings.TrimSpace(in.SupplierSKU),
		CostHintMinor:    in.CostHintMinor,
		CostHintCurrency: strings.ToUpper(strings.TrimSpace(in.CostHintCurrency)),
		LeadTimeDays:     in.LeadTimeDays,
		MOQ:              in.MOQ,
		Metadata:         map[string]any{},
		IsPreferred:      in.IsPreferred,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := sp.Validate(); err != nil {
		return domain.SupplierProduct{}, err
	}
	if err := d.Suppliers.LinkProduct(ctx, sp); err != nil {
		return domain.SupplierProduct{}, err
	}
	d.publishEvent(ctx, domain.EventSupplierChanged, in.TenantID, in.ProductID, map[string]any{
		"supplierId": in.SupplierID, "productId": in.ProductID, "action": "linked",
	})
	return sp, nil
}

// ListSupplierProducts lists product links for a supplier.
func (d *Deps) ListSupplierProducts(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierProduct, error) {
	if d.Suppliers == nil {
		return nil, nil
	}
	return d.Suppliers.ListProducts(ctx, tenantID, supplierID)
}
