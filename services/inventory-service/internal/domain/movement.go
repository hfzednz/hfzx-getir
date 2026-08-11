package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MovementType classifies a ledger entry.
type MovementType string

const (
	MovementTypeReceipt         MovementType = "receipt"
	MovementTypePurchaseReceipt MovementType = "purchase_receipt"
	MovementTypeSale            MovementType = "sale"
	MovementTypeCourierPickup   MovementType = "courier_pickup"
	MovementTypeTransferOut     MovementType = "transfer_out"
	MovementTypeTransferIn      MovementType = "transfer_in"
	MovementTypeAdjust          MovementType = "adjust"
	MovementTypeCount           MovementType = "count"
	MovementTypeDamage          MovementType = "damage"
	MovementTypeReturnIn        MovementType = "return_in"
	MovementTypeSupplierReturn  MovementType = "supplier_return"
	MovementTypeWaste           MovementType = "waste"
	MovementTypeManual          MovementType = "manual"
)

func (t MovementType) Valid() bool {
	switch t {
	case MovementTypeReceipt, MovementTypePurchaseReceipt, MovementTypeSale,
		MovementTypeCourierPickup, MovementTypeTransferOut, MovementTypeTransferIn,
		MovementTypeAdjust, MovementTypeCount, MovementTypeDamage,
		MovementTypeReturnIn, MovementTypeSupplierReturn, MovementTypeWaste,
		MovementTypeManual:
		return true
	default:
		return false
	}
}

// Movement is an append-only stock ledger entry with balance-after snapshot.
type Movement struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	BalanceID      *uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LocationID     *uuid.UUID
	LotID          *uuid.UUID
	Type           MovementType
	Qty            int64 // signed delta
	OnHandAfter    int64
	ReservedAfter  int64
	BlockedAfter   int64
	IncomingAfter  int64
	InTransitAfter int64
	IdempotencyKey string
	ActorID        *uuid.UUID
	Reason         string
	ExternalRef    string
	ReservationID  *uuid.UUID
	TransferID     *uuid.UUID
	Metadata       map[string]any
	OccurredAt     time.Time
	CreatedAt      time.Time
}

// Validate checks structural invariants.
func (m Movement) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: movement id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if m.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if m.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if !m.Type.Valid() {
		return fmt.Errorf("%w: invalid movement type %q", ErrInvalidArgument, m.Type)
	}
	if m.Qty == 0 {
		return fmt.Errorf("%w: movement qty cannot be zero", ErrInvalidArgument)
	}
	if m.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	if m.OnHandAfter < 0 || m.ReservedAfter < 0 || m.BlockedAfter < 0 ||
		m.IncomingAfter < 0 || m.InTransitAfter < 0 {
		return fmt.Errorf("%w: snapshot buckets cannot be negative", ErrInvariant)
	}
	if m.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidArgument)
	}
	return nil
}

// NewMovementFromBalance builds a movement with snapshot fields from balance.
func NewMovementFromBalance(
	id uuid.UUID,
	tenantID uuid.UUID,
	balance StockBalance,
	movType MovementType,
	qty int64,
	idempotencyKey string,
	actorID *uuid.UUID,
	reason string,
) (Movement, error) {
	m := Movement{
		ID:             id,
		TenantID:       tenantID,
		WarehouseID:    balance.WarehouseID,
		BalanceID:      &balance.ID,
		VariantID:      balance.VariantID,
		SKUCode:        balance.SKUCode,
		LocationID:     balance.LocationID,
		Type:           movType,
		Qty:            qty,
		OnHandAfter:    balance.OnHand,
		ReservedAfter:  balance.Reserved,
		BlockedAfter:   balance.Blocked,
		IncomingAfter:  balance.Incoming,
		InTransitAfter: balance.InTransit,
		IdempotencyKey: idempotencyKey,
		ActorID:        actorID,
		Reason:         reason,
		OccurredAt:     time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
	}
	if err := m.Validate(); err != nil {
		return Movement{}, err
	}
	return m, nil
}
