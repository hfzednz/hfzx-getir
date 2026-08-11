package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// QueryATP computes availability-to-promise for a variant (optional warehouse/region).
func (d *Deps) QueryATP(ctx context.Context, q domain.ATPQuery) ([]domain.ATPResult, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	asOf := q.AsOf
	if asOf.IsZero() {
		asOf = d.now()
	}

	var balances []domain.StockBalance
	if q.VariantID != uuid.Nil {
		all, err := d.Balances.ListByVariant(ctx, q.TenantID, q.VariantID)
		if err != nil {
			return nil, err
		}
		balances = all
	}
	if q.WarehouseID != nil && q.VariantID != uuid.Nil {
		filtered := balances[:0]
		for _, b := range balances {
			if b.WarehouseID == *q.WarehouseID {
				filtered = append(filtered, b)
			}
		}
		balances = filtered
		if len(balances) == 0 {
			if bal, err := d.Balances.GetByKey(ctx, ports.BalanceKey{
				TenantID: q.TenantID, WarehouseID: *q.WarehouseID,
				VariantID: q.VariantID, LocationID: nil,
			}); err == nil {
				balances = []domain.StockBalance{bal}
			}
		}
	}

	out := make([]domain.ATPResult, 0, len(balances))
	for _, b := range balances {
		if q.SKUCode != "" && b.SKUCode != "" && b.SKUCode != q.SKUCode {
			continue
		}
		if q.RegionID != nil {
			wh, err := d.Warehouses.GetByID(ctx, q.TenantID, b.WarehouseID)
			if err != nil {
				continue
			}
			if wh.RegionID == nil || *wh.RegionID != *q.RegionID {
				continue
			}
		}
		out = append(out, domain.ComputeATP(b, q.Policy, asOf))
	}
	return out, nil
}
