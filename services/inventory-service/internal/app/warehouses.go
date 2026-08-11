package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// CreateWarehouseInput creates a warehouse.
type CreateWarehouseInput struct {
	TenantID uuid.UUID
	Code     string
	Name     string
	RegionID *uuid.UUID
	Timezone string
	Status   domain.WarehouseStatus
}

// CreateWarehouse persists a new warehouse.
func (d *Deps) CreateWarehouse(ctx context.Context, in CreateWarehouseInput) (domain.Warehouse, error) {
	if in.TenantID == uuid.Nil {
		return domain.Warehouse{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	status := in.Status
	if status == "" {
		status = domain.WarehouseStatusActive
	}
	now := d.now()
	w := domain.Warehouse{
		ID:        d.newID(),
		TenantID:  in.TenantID,
		Code:      code,
		Name:      name,
		RegionID:  in.RegionID,
		Timezone:  tz,
		Status:    status,
		Metadata:  map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := w.Validate(); err != nil {
		return domain.Warehouse{}, err
	}
	if _, err := d.Warehouses.GetByCode(ctx, in.TenantID, code); err == nil {
		return domain.Warehouse{}, domain.ErrAlreadyExists
	}
	if err := d.Warehouses.Create(ctx, w); err != nil {
		return domain.Warehouse{}, err
	}
	return w, nil
}

// UpdateWarehouseInput patches warehouse fields.
type UpdateWarehouseInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	Name     *string
	RegionID *uuid.UUID
	Timezone *string
	Status   *domain.WarehouseStatus
}

// UpdateWarehouse updates an existing warehouse.
func (d *Deps) UpdateWarehouse(ctx context.Context, in UpdateWarehouseInput) (domain.Warehouse, error) {
	w, err := d.Warehouses.GetByID(ctx, in.TenantID, in.ID)
	if err != nil {
		return domain.Warehouse{}, err
	}
	if in.Name != nil {
		w.Name = strings.TrimSpace(*in.Name)
	}
	if in.RegionID != nil {
		w.RegionID = in.RegionID
	}
	if in.Timezone != nil {
		w.Timezone = strings.TrimSpace(*in.Timezone)
	}
	if in.Status != nil {
		w.Status = *in.Status
	}
	w.UpdatedAt = d.now()
	if err := w.Validate(); err != nil {
		return domain.Warehouse{}, err
	}
	if err := d.Warehouses.Update(ctx, w); err != nil {
		return domain.Warehouse{}, err
	}
	return w, nil
}

// GetWarehouse returns a warehouse by id.
func (d *Deps) GetWarehouse(ctx context.Context, tenantID, id uuid.UUID) (domain.Warehouse, error) {
	return d.Warehouses.GetByID(ctx, tenantID, id)
}

// ListWarehouses lists warehouses for a tenant.
func (d *Deps) ListWarehouses(ctx context.Context, f ports.WarehouseFilter) ([]domain.Warehouse, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	return d.Warehouses.List(ctx, f)
}

// DeleteWarehouse soft-deletes a warehouse.
func (d *Deps) DeleteWarehouse(ctx context.Context, tenantID, id uuid.UUID) error {
	return d.Warehouses.Delete(ctx, tenantID, id, d.now())
}
