package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// CreateLocationInput creates a location node.
type CreateLocationInput struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	ParentID    *uuid.UUID
	Kind        domain.LocationKind
	ZoneType    *domain.ZoneType
	Code        string
	Name        string
	IsPickable  bool
}

// CreateLocation persists a location under a warehouse.
func (d *Deps) CreateLocation(ctx context.Context, in CreateLocationInput) (domain.Location, error) {
	wh, err := d.Warehouses.GetByID(ctx, in.TenantID, in.WarehouseID)
	if err != nil {
		return domain.Location{}, err
	}
	if !wh.IsOperable() {
		return domain.Location{}, fmt.Errorf("%w: warehouse not operable", domain.ErrInvariant)
	}
	parentPath := "/"
	depth := 0
	if in.ParentID != nil {
		parent, err := d.Locations.GetByID(ctx, *in.ParentID)
		if err != nil {
			return domain.Location{}, err
		}
		if parent.WarehouseID != in.WarehouseID {
			return domain.Location{}, fmt.Errorf("%w: parent warehouse mismatch", domain.ErrInvariant)
		}
		parentPath = parent.Path
		depth = parent.Depth + 1
	}
	id := d.newID()
	now := d.now()
	loc := domain.Location{
		ID:          id,
		WarehouseID: in.WarehouseID,
		ParentID:    in.ParentID,
		Kind:        in.Kind,
		ZoneType:    in.ZoneType,
		Code:        strings.TrimSpace(in.Code),
		Path:        domain.BuildPath(parentPath, id),
		Depth:       depth,
		Name:        strings.TrimSpace(in.Name),
		IsPickable:  in.IsPickable,
		IsActive:    true,
		Metadata:    map[string]any{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := loc.Validate(); err != nil {
		return domain.Location{}, err
	}
	if err := d.Locations.Create(ctx, loc); err != nil {
		return domain.Location{}, err
	}
	return loc, nil
}

// UpdateLocationInput patches a location.
type UpdateLocationInput struct {
	ID         uuid.UUID
	Name       *string
	ZoneType   *domain.ZoneType
	IsPickable *bool
	IsActive   *bool
}

// UpdateLocation updates a location node.
func (d *Deps) UpdateLocation(ctx context.Context, in UpdateLocationInput) (domain.Location, error) {
	loc, err := d.Locations.GetByID(ctx, in.ID)
	if err != nil {
		return domain.Location{}, err
	}
	if in.Name != nil {
		loc.Name = strings.TrimSpace(*in.Name)
	}
	if in.ZoneType != nil {
		loc.ZoneType = in.ZoneType
	}
	if in.IsPickable != nil {
		loc.IsPickable = *in.IsPickable
	}
	if in.IsActive != nil {
		loc.IsActive = *in.IsActive
	}
	loc.UpdatedAt = d.now()
	if err := loc.Validate(); err != nil {
		return domain.Location{}, err
	}
	if err := d.Locations.Update(ctx, loc); err != nil {
		return domain.Location{}, err
	}
	return loc, nil
}

// GetLocation returns a location by id.
func (d *Deps) GetLocation(ctx context.Context, id uuid.UUID) (domain.Location, error) {
	return d.Locations.GetByID(ctx, id)
}

// ListLocations lists locations for a warehouse.
func (d *Deps) ListLocations(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error) {
	return d.Locations.ListByWarehouse(ctx, warehouseID)
}

// DeleteLocation soft-deletes a location.
func (d *Deps) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	return d.Locations.Delete(ctx, id, d.now())
}
