package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// ReceiveReturnLine is a return line.
type ReceiveReturnLine struct {
	VariantID   uuid.UUID
	SKUCode     string
	LotID       *uuid.UUID
	LocationID  *uuid.UUID
	Qty         int64
	Disposition *domain.ReturnDisposition
}

// ReceiveReturnCmd receives returned stock into inventory.
type ReceiveReturnCmd struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	Source         domain.ReturnSource
	Disposition    domain.ReturnDisposition
	ExternalRef    string
	ActorID        *uuid.UUID
	Reason         string
	Lines          []ReceiveReturnLine
	IdempotencyKey string
}

// ReceiveReturn posts return stock per disposition (restock/quarantine/waste).
func (d *Deps) ReceiveReturn(ctx context.Context, in ReceiveReturnCmd) (domain.InventoryReturn, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || len(in.Lines) == 0 {
		return domain.InventoryReturn{}, domain.ErrInvalidArgument
	}
	if in.IdempotencyKey == "" {
		return domain.InventoryReturn{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "return:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.InventoryReturn); ok {
			return r, nil
		}
	}
	disp := in.Disposition
	if disp == "" {
		disp = domain.ReturnDispositionRestock
	}
	src := in.Source
	if src == "" {
		src = domain.ReturnSourceCustomer
	}
	now := d.now()
	ret := domain.InventoryReturn{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		Source: src, Disposition: disp, Status: domain.ReturnStatusDraft,
		ExternalRef: in.ExternalRef, ActorID: in.ActorID, Reason: in.Reason,
		Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	for _, line := range in.Lines {
		ret.Lines = append(ret.Lines, domain.ReturnLine{
			ID: d.newID(), ReturnID: ret.ID, VariantID: line.VariantID,
			SKUCode: line.SKUCode, LotID: line.LotID, LocationID: line.LocationID,
			Qty: line.Qty, Disposition: line.Disposition,
			Metadata: map[string]any{}, CreatedAt: now,
		})
	}
	if err := ret.Validate(); err != nil {
		return domain.InventoryReturn{}, err
	}
	if err := d.Returns.Create(ctx, ret); err != nil {
		return domain.InventoryReturn{}, err
	}
	if err := ret.TransitionTo(domain.ReturnStatusReceived); err != nil {
		return domain.InventoryReturn{}, err
	}

	for i, line := range ret.Lines {
		eff := line.EffectiveDisposition(disp)
		switch eff {
		case domain.ReturnDispositionRestock:
			_, _, err := d.Receive(ctx, ReceiveStockCmd{
				TenantID: in.TenantID, WarehouseID: in.WarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: line.LocationID,
				Qty: line.Qty, IdempotencyKey: in.IdempotencyKey + ":line:" + line.ID.String(),
				ActorID: in.ActorID, Reason: "return restock",
			})
			if err != nil {
				return domain.InventoryReturn{}, err
			}
		case domain.ReturnDispositionQuarantine:
			b, err := d.EnsureBalance(ctx, EnsureBalanceInput{
				TenantID: in.TenantID, WarehouseID: in.WarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: line.LocationID,
			})
			if err != nil {
				return domain.InventoryReturn{}, err
			}
			if err := b.AdjustOnHand(line.Qty); err != nil {
				return domain.InventoryReturn{}, err
			}
			if err := b.Block(line.Qty); err != nil {
				return domain.InventoryReturn{}, err
			}
			if err := d.Balances.Update(ctx, b); err != nil {
				return domain.InventoryReturn{}, err
			}
			d.indexBalance(ctx, b)
		case domain.ReturnDispositionWaste:
			// Received then immediately written off — no net on_hand change.
		}
		_ = i
	}
	if err := ret.TransitionTo(domain.ReturnStatusDisposed); err != nil {
		return domain.InventoryReturn{}, err
	}
	if err := d.Returns.Update(ctx, ret); err != nil {
		return domain.InventoryReturn{}, err
	}
	d.idemPut(ctx, "return:"+in.IdempotencyKey, ret)
	return ret, nil
}

// GetReturn returns an inventory return.
func (d *Deps) GetReturn(ctx context.Context, tenantID, id uuid.UUID) (domain.InventoryReturn, error) {
	return d.Returns.GetByID(ctx, tenantID, id)
}
