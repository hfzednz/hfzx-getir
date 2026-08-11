package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// ListNearExpiry returns lots expiring within the given day window.
func (d *Deps) ListNearExpiry(ctx context.Context, tenantID uuid.UUID, warehouseID *uuid.UUID, withinDays int) ([]domain.Lot, error) {
	if withinDays <= 0 {
		withinDays = 7
	}
	return d.Lots.ListNearExpiry(ctx, tenantID, warehouseID, withinDays, d.now())
}

// AllocateFEFOCmd requests FEFO allocation without reserving.
type AllocateFEFOCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	Qty         int64
	AsOf        time.Time
}

// AllocateFEFO returns FEFO lot allocations (helper used by SoftReserve when UseFEFO).
func (d *Deps) AllocateFEFO(ctx context.Context, in AllocateFEFOCmd) ([]domain.LotAllocation, error) {
	asOf := in.AsOf
	if asOf.IsZero() {
		asOf = d.now()
	}
	lots, err := d.Lots.ListByWarehouseVariant(ctx, in.TenantID, in.WarehouseID, in.VariantID)
	if err != nil {
		return nil, err
	}
	return domain.AllocateLotsFEFO(lots, in.Qty, asOf)
}
