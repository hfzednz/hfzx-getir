package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StockBalance is the quantity SoT for (warehouse, variant, optional location).
// VariantID / SKUCode are opaque catalog references — no product content.
type StockBalance struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	WarehouseID  uuid.UUID
	VariantID    uuid.UUID
	SKUCode      string
	LocationID   *uuid.UUID
	OnHand       int64
	Reserved     int64
	Blocked      int64
	Incoming     int64
	InTransit    int64
	Version      int64
	SafetyMin    int64
	ReorderPoint int64
	MaxStock     *int64
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Available returns sellable / reservable quantity: on_hand - reserved - blocked.
func (b StockBalance) Available() int64 {
	avail := b.OnHand - b.Reserved - b.Blocked
	if avail < 0 {
		return 0
	}
	return avail
}

// Validate checks structural invariants.
func (b StockBalance) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: balance id required", ErrInvalidArgument)
	}
	if b.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if b.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if b.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if b.OnHand < 0 || b.Reserved < 0 || b.Blocked < 0 || b.Incoming < 0 || b.InTransit < 0 {
		return fmt.Errorf("%w: quantity buckets cannot be negative", ErrInvariant)
	}
	if b.OnHand < b.Reserved+b.Blocked {
		return fmt.Errorf("%w: on_hand < reserved + blocked", ErrInvariant)
	}
	if b.Version < 1 {
		return fmt.Errorf("%w: version must be >= 1", ErrInvariant)
	}
	if b.SafetyMin < 0 || b.ReorderPoint < 0 {
		return fmt.Errorf("%w: safety/reorder thresholds cannot be negative", ErrInvalidArgument)
	}
	if b.MaxStock != nil && *b.MaxStock < 0 {
		return fmt.Errorf("%w: max_stock cannot be negative", ErrInvalidArgument)
	}
	return nil
}

// bumpVersion increments the optimistic concurrency token.
func (b *StockBalance) bumpVersion() {
	b.Version++
	b.UpdatedAt = time.Now().UTC()
}

// ExpectVersion returns ErrVersionConflict when expected does not match.
func (b StockBalance) ExpectVersion(expected int64) error {
	if b.Version != expected {
		return fmt.Errorf("%w: expected %d got %d", ErrVersionConflict, expected, b.Version)
	}
	return nil
}

// Reserve holds qty against Available(). Invariant: cannot reserve more than Available().
func (b *StockBalance) Reserve(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: reserve qty", ErrNegativeQuantity)
	}
	if qty > b.Available() {
		return fmt.Errorf("%w: requested %d available %d", ErrInsufficientStock, qty, b.Available())
	}
	b.Reserved += qty
	b.bumpVersion()
	return nil
}

// ReleaseReserved frees a previously reserved quantity.
func (b *StockBalance) ReleaseReserved(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: release qty", ErrNegativeQuantity)
	}
	if qty > b.Reserved {
		return fmt.Errorf("%w: release %d exceeds reserved %d", ErrInvariant, qty, b.Reserved)
	}
	b.Reserved -= qty
	b.bumpVersion()
	return nil
}

// ConsumeReserved converts a hard reservation into an on_hand deduction (ship/deduct).
func (b *StockBalance) ConsumeReserved(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: consume qty", ErrNegativeQuantity)
	}
	if qty > b.Reserved {
		return fmt.Errorf("%w: consume %d exceeds reserved %d", ErrInvariant, qty, b.Reserved)
	}
	if qty > b.OnHand {
		return fmt.Errorf("%w: consume %d exceeds on_hand %d", ErrInvariant, qty, b.OnHand)
	}
	b.Reserved -= qty
	b.OnHand -= qty
	b.bumpVersion()
	return nil
}

// AdjustOnHand applies a signed delta to on_hand (adjust/count/damage/waste/receipt).
func (b *StockBalance) AdjustOnHand(delta int64) error {
	next := b.OnHand + delta
	if next < 0 {
		return fmt.Errorf("%w: on_hand would become negative", ErrInvariant)
	}
	if next < b.Reserved+b.Blocked {
		return fmt.Errorf("%w: adjust would violate on_hand >= reserved + blocked", ErrInvariant)
	}
	b.OnHand = next
	b.bumpVersion()
	return nil
}

// Block marks qty as blocked (quarantine / hold).
func (b *StockBalance) Block(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: block qty", ErrNegativeQuantity)
	}
	if qty > b.Available() {
		return fmt.Errorf("%w: block %d available %d", ErrInsufficientStock, qty, b.Available())
	}
	b.Blocked += qty
	b.bumpVersion()
	return nil
}

// Unblock releases blocked quantity back to available.
func (b *StockBalance) Unblock(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: unblock qty", ErrNegativeQuantity)
	}
	if qty > b.Blocked {
		return fmt.Errorf("%w: unblock %d exceeds blocked %d", ErrInvariant, qty, b.Blocked)
	}
	b.Blocked -= qty
	b.bumpVersion()
	return nil
}

// SnapshotAfter returns post-mutation bucket values for movement ledger rows.
func (b StockBalance) SnapshotAfter() (onHand, reserved, blocked, incoming, inTransit int64) {
	return b.OnHand, b.Reserved, b.Blocked, b.Incoming, b.InTransit
}

// NeedsReorder reports whether on_hand is at or below reorder_point.
func (b StockBalance) NeedsReorder() bool {
	return b.OnHand <= b.ReorderPoint
}
