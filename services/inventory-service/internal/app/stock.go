package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// EnsureBalanceInput creates a zero balance if missing.
type EnsureBalanceInput struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	LocationID  *uuid.UUID
}

// EnsureBalance returns an existing balance or creates a zero one.
func (d *Deps) EnsureBalance(ctx context.Context, in EnsureBalanceInput) (domain.StockBalance, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.VariantID == uuid.Nil {
		return domain.StockBalance{}, domain.ErrInvalidArgument
	}
	key := ports.BalanceKey{
		TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		VariantID: in.VariantID, LocationID: in.LocationID,
	}
	if b, err := d.Balances.GetByKey(ctx, key); err == nil {
		return b, nil
	}
	now := d.now()
	b := domain.StockBalance{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		WarehouseID: in.WarehouseID,
		VariantID:   in.VariantID,
		SKUCode:     in.SKUCode,
		LocationID:  in.LocationID,
		Version:     1,
		Metadata:    map[string]any{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := b.Validate(); err != nil {
		return domain.StockBalance{}, err
	}
	if err := d.Balances.Create(ctx, b); err != nil {
		// race: another writer created it
		if existing, err2 := d.Balances.GetByKey(ctx, key); err2 == nil {
			return existing, nil
		}
		return domain.StockBalance{}, err
	}
	d.publishEvent(ctx, domain.EventInventoryCreated, b.TenantID, b.WarehouseID, b.VariantID, map[string]any{
		"balanceId": b.ID,
	})
	d.indexBalance(ctx, b)
	return b, nil
}

// GetBalance returns a balance by id.
func (d *Deps) GetBalance(ctx context.Context, tenantID, id uuid.UUID) (domain.StockBalance, error) {
	return d.Balances.GetByID(ctx, tenantID, id)
}

// ListBalances lists warehouse balances.
func (d *Deps) ListBalances(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockBalance, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Balances.ListByWarehouse(ctx, tenantID, warehouseID, limit, offset)
}

// AdjustStockInput adjusts on_hand by a signed delta.
type AdjustStockInput struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LocationID     *uuid.UUID
	Delta          int64
	IdempotencyKey string
	ActorID        *uuid.UUID
	Reason         string
}

// Adjust applies a signed on_hand delta (idempotent).
func (d *Deps) Adjust(ctx context.Context, in AdjustStockInput) (domain.StockBalance, domain.Movement, error) {
	if in.Delta == 0 {
		return domain.StockBalance{}, domain.Movement{}, domain.ErrInvalidArgument
	}
	if in.IdempotencyKey == "" {
		return domain.StockBalance{}, domain.Movement{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "adjust:"+in.IdempotencyKey); ok {
		if pair, ok := v.(stockResult); ok {
			return pair.Balance, pair.Movement, nil
		}
	}
	if m, err := d.Movements.GetByIdempotencyKey(ctx, in.IdempotencyKey); err == nil {
		b, err2 := d.Balances.GetByID(ctx, in.TenantID, *m.BalanceID)
		if err2 != nil {
			return domain.StockBalance{}, m, err2
		}
		return b, m, nil
	}

	var outB domain.StockBalance
	var outM domain.Movement
	err := d.withStockLock(ctx, stockKey(in.WarehouseID, in.VariantID, in.LocationID), func() error {
		b, err := d.EnsureBalance(ctx, EnsureBalanceInput{
			TenantID: in.TenantID, WarehouseID: in.WarehouseID,
			VariantID: in.VariantID, SKUCode: in.SKUCode, LocationID: in.LocationID,
		})
		if err != nil {
			return err
		}
		if err := b.AdjustOnHand(in.Delta); err != nil {
			return err
		}
		if err := d.Balances.Update(ctx, b); err != nil {
			return err
		}
		m, err := domain.NewMovementFromBalance(d.newID(), in.TenantID, b, domain.MovementTypeAdjust, in.Delta, in.IdempotencyKey, in.ActorID, in.Reason)
		if err != nil {
			return err
		}
		if err := d.Movements.Create(ctx, m); err != nil {
			return err
		}
		outB, outM = b, m
		d.publishEvent(ctx, domain.EventStockAdjusted, b.TenantID, b.WarehouseID, b.VariantID, map[string]any{
			"balanceId": b.ID, "delta": in.Delta, "movementId": m.ID,
		})
		d.indexBalance(ctx, b)
		return nil
	})
	if err != nil {
		return domain.StockBalance{}, domain.Movement{}, err
	}
	d.idemPut(ctx, "adjust:"+in.IdempotencyKey, stockResult{Balance: outB, Movement: outM})
	return outB, outM, nil
}

type stockResult struct {
	Balance  domain.StockBalance
	Movement domain.Movement
}

// ReceiveStockCmd receives inbound stock (positive qty).
type ReceiveStockCmd struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LocationID     *uuid.UUID
	Qty            int64
	LotCode        string
	IdempotencyKey string
	ActorID        *uuid.UUID
	Reason         string
}

// Receive increases on_hand (idempotent).
func (d *Deps) Receive(ctx context.Context, in ReceiveStockCmd) (domain.StockBalance, domain.Movement, error) {
	if in.Qty <= 0 {
		return domain.StockBalance{}, domain.Movement{}, domain.ErrNegativeQuantity
	}
	if in.IdempotencyKey == "" {
		return domain.StockBalance{}, domain.Movement{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "receive:"+in.IdempotencyKey); ok {
		if pair, ok := v.(stockResult); ok {
			return pair.Balance, pair.Movement, nil
		}
	}
	if m, err := d.Movements.GetByIdempotencyKey(ctx, in.IdempotencyKey); err == nil {
		b, err2 := d.Balances.GetByID(ctx, in.TenantID, *m.BalanceID)
		if err2 != nil {
			return domain.StockBalance{}, m, err2
		}
		return b, m, nil
	}

	var outB domain.StockBalance
	var outM domain.Movement
	err := d.withStockLock(ctx, stockKey(in.WarehouseID, in.VariantID, in.LocationID), func() error {
		b, err := d.EnsureBalance(ctx, EnsureBalanceInput{
			TenantID: in.TenantID, WarehouseID: in.WarehouseID,
			VariantID: in.VariantID, SKUCode: in.SKUCode, LocationID: in.LocationID,
		})
		if err != nil {
			return err
		}
		if err := b.AdjustOnHand(in.Qty); err != nil {
			return err
		}
		if err := d.Balances.Update(ctx, b); err != nil {
			return err
		}
		if in.LotCode != "" {
			now := d.now()
			lot := domain.Lot{
				ID: d.newID(), TenantID: in.TenantID, BalanceID: b.ID,
				WarehouseID: in.WarehouseID, VariantID: in.VariantID,
				LotCode: in.LotCode, Qty: in.Qty, Status: domain.LotStatusGood,
				ReceivedAt: &now, Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
			}
			if err := lot.Validate(); err != nil {
				return err
			}
			if err := d.Lots.Create(ctx, lot); err != nil {
				return err
			}
		}
		m, err := domain.NewMovementFromBalance(d.newID(), in.TenantID, b, domain.MovementTypeReceipt, in.Qty, in.IdempotencyKey, in.ActorID, in.Reason)
		if err != nil {
			return err
		}
		if err := d.Movements.Create(ctx, m); err != nil {
			return err
		}
		outB, outM = b, m
		d.publishEvent(ctx, domain.EventStockReceived, b.TenantID, b.WarehouseID, b.VariantID, map[string]any{
			"balanceId": b.ID, "qty": in.Qty, "movementId": m.ID,
		})
		d.indexBalance(ctx, b)
		return nil
	})
	if err != nil {
		return domain.StockBalance{}, domain.Movement{}, err
	}
	d.idemPut(ctx, "receive:"+in.IdempotencyKey, stockResult{Balance: outB, Movement: outM})
	return outB, outM, nil
}

// DamageStockCmd writes off damaged qty from available.
type DamageStockCmd struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LocationID     *uuid.UUID
	Qty            int64
	IdempotencyKey string
	ActorID        *uuid.UUID
	Reason         string
}

// Damage decreases on_hand for damaged stock.
func (d *Deps) Damage(ctx context.Context, in DamageStockCmd) (domain.StockBalance, domain.Movement, error) {
	return d.mutateDown(ctx, "damage:", in.TenantID, in.WarehouseID, in.VariantID, in.SKUCode, in.LocationID, in.Qty, in.IdempotencyKey, in.ActorID, in.Reason, domain.MovementTypeDamage)
}

// WasteStockCmd writes off waste qty.
type WasteStockCmd struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LocationID     *uuid.UUID
	Qty            int64
	IdempotencyKey string
	ActorID        *uuid.UUID
	Reason         string
}

// Waste decreases on_hand for waste.
func (d *Deps) Waste(ctx context.Context, in WasteStockCmd) (domain.StockBalance, domain.Movement, error) {
	return d.mutateDown(ctx, "waste:", in.TenantID, in.WarehouseID, in.VariantID, in.SKUCode, in.LocationID, in.Qty, in.IdempotencyKey, in.ActorID, in.Reason, domain.MovementTypeWaste)
}

func (d *Deps) mutateDown(
	ctx context.Context, prefix string,
	tenantID, warehouseID, variantID uuid.UUID, sku string, locationID *uuid.UUID,
	qty int64, idemKey string, actorID *uuid.UUID, reason string, movType domain.MovementType,
) (domain.StockBalance, domain.Movement, error) {
	if qty <= 0 {
		return domain.StockBalance{}, domain.Movement{}, domain.ErrNegativeQuantity
	}
	if idemKey == "" {
		return domain.StockBalance{}, domain.Movement{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, prefix+idemKey); ok {
		if pair, ok := v.(stockResult); ok {
			return pair.Balance, pair.Movement, nil
		}
	}
	if m, err := d.Movements.GetByIdempotencyKey(ctx, idemKey); err == nil {
		b, err2 := d.Balances.GetByID(ctx, tenantID, *m.BalanceID)
		if err2 != nil {
			return domain.StockBalance{}, m, err2
		}
		return b, m, nil
	}

	var outB domain.StockBalance
	var outM domain.Movement
	err := d.withStockLock(ctx, stockKey(warehouseID, variantID, locationID), func() error {
		b, err := d.EnsureBalance(ctx, EnsureBalanceInput{
			TenantID: tenantID, WarehouseID: warehouseID,
			VariantID: variantID, SKUCode: sku, LocationID: locationID,
		})
		if err != nil {
			return err
		}
		if qty > b.Available() {
			return fmt.Errorf("%w: requested %d available %d", domain.ErrInsufficientStock, qty, b.Available())
		}
		if err := b.AdjustOnHand(-qty); err != nil {
			return err
		}
		if err := d.Balances.Update(ctx, b); err != nil {
			return err
		}
		m, err := domain.NewMovementFromBalance(d.newID(), tenantID, b, movType, -qty, idemKey, actorID, reason)
		if err != nil {
			return err
		}
		if err := d.Movements.Create(ctx, m); err != nil {
			return err
		}
		outB, outM = b, m
		d.publishEvent(ctx, domain.EventStockAdjusted, b.TenantID, b.WarehouseID, b.VariantID, map[string]any{
			"balanceId": b.ID, "delta": -qty, "type": string(movType), "movementId": m.ID,
		})
		d.indexBalance(ctx, b)
		return nil
	})
	if err != nil {
		return domain.StockBalance{}, domain.Movement{}, err
	}
	d.idemPut(ctx, prefix+idemKey, stockResult{Balance: outB, Movement: outM})
	return outB, outM, nil
}

// ListMovements lists ledger rows for a balance.
func (d *Deps) ListMovements(ctx context.Context, balanceID uuid.UUID, limit, offset int) ([]domain.Movement, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Movements.ListByBalance(ctx, balanceID, limit, offset)
}
