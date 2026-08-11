package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReservationType distinguishes soft (cart) vs hard (order) holds.
type ReservationType string

const (
	ReservationTypeSoft ReservationType = "soft"
	ReservationTypeHard ReservationType = "hard"
)

func (t ReservationType) Valid() bool {
	switch t {
	case ReservationTypeSoft, ReservationTypeHard:
		return true
	default:
		return false
	}
}

// ReservationStatus is the lifecycle of a reservation header.
type ReservationStatus string

const (
	ReservationStatusActive   ReservationStatus = "active"
	ReservationStatusReleased ReservationStatus = "released"
	ReservationStatusConsumed ReservationStatus = "consumed"
	ReservationStatusExpired  ReservationStatus = "expired"
)

func (s ReservationStatus) Valid() bool {
	switch s {
	case ReservationStatusActive, ReservationStatusReleased,
		ReservationStatusConsumed, ReservationStatusExpired:
		return true
	default:
		return false
	}
}

// Allowed transitions:
//
//	Soft  → Hard (confirm), Soft → Released, Soft → Expired
//	Hard  → Consumed, Hard → Released
var reservationTransitions = map[ReservationType]map[ReservationStatus][]ReservationStatus{
	ReservationTypeSoft: {
		ReservationStatusActive: {
			ReservationStatusReleased,
			ReservationStatusExpired,
		},
	},
	ReservationTypeHard: {
		ReservationStatusActive: {
			ReservationStatusConsumed,
			ReservationStatusReleased,
		},
	},
}

// Reservation is a soft/hard stock hold header (multi-warehouse via lines).
// ExternalRef is an opaque cart/order id — no order aggregate here.
type Reservation struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID *uuid.UUID
	Type        ReservationType
	Status      ReservationStatus
	ExpiresAt   *time.Time
	Priority    int
	ExternalRef string
	ActorID     *uuid.UUID
	Lines       []ReservationLine
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ReleasedAt  *time.Time
	ConsumedAt  *time.Time
}

// ReservationLine is a per-warehouse/variant quantity hold.
type ReservationLine struct {
	ID            uuid.UUID
	ReservationID uuid.UUID
	WarehouseID   uuid.UUID
	VariantID     uuid.UUID
	SKUCode       string
	Qty           int64
	LotID         *uuid.UUID
	BalanceID     *uuid.UUID
	LocationID    *uuid.UUID
	Metadata      map[string]any
	CreatedAt     time.Time
}

// Validate checks structural invariants.
func (r Reservation) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: reservation id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !r.Type.Valid() {
		return fmt.Errorf("%w: invalid reservation type %q", ErrInvalidArgument, r.Type)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid reservation status %q", ErrInvalidArgument, r.Status)
	}
	if r.Type == ReservationTypeSoft && r.Status == ReservationStatusActive && r.ExpiresAt == nil {
		return fmt.Errorf("%w: soft reservation requires expires_at", ErrInvariant)
	}
	for i, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks a reservation line.
func (l ReservationLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: reservation line id required", ErrInvalidArgument)
	}
	if l.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: line qty", ErrNegativeQuantity)
	}
	return nil
}

// IsActive reports whether the reservation still holds stock.
func (r Reservation) IsActive() bool {
	return r.Status == ReservationStatusActive
}

// IsExpiredAt reports whether an active soft reservation has passed expires_at.
func (r Reservation) IsExpiredAt(asOf time.Time) bool {
	if r.Status != ReservationStatusActive || r.ExpiresAt == nil {
		return false
	}
	return !asOf.Before(*r.ExpiresAt)
}

// CanTransitionTo reports whether type/status allows moving to next.
// Soft→Hard is a type change handled by ConfirmHard, not a status-only transition.
func (r Reservation) CanTransitionTo(next ReservationStatus) bool {
	if !r.Status.Valid() || !next.Valid() {
		return false
	}
	if r.Status == next {
		return true
	}
	allowed, ok := reservationTransitions[r.Type]
	if !ok {
		return false
	}
	for _, s := range allowed[r.Status] {
		if s == next {
			return true
		}
	}
	return false
}

func (r *Reservation) applyStatus(next ReservationStatus) error {
	if !r.IsActive() {
		return fmt.Errorf("%w: status %s", ErrReservationInactive, r.Status)
	}
	if !r.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s/%s → %s", ErrInvalidTransition, r.Type, r.Status, next)
	}
	now := time.Now().UTC()
	r.Status = next
	r.UpdatedAt = now
	switch next {
	case ReservationStatusReleased, ReservationStatusExpired:
		r.ReleasedAt = &now
	case ReservationStatusConsumed:
		r.ConsumedAt = &now
	}
	return nil
}

// ConfirmHard transitions Soft → Hard (checkout). Soft→Hard only.
func (r *Reservation) ConfirmHard() error {
	if !r.IsActive() {
		return fmt.Errorf("%w: status %s", ErrReservationInactive, r.Status)
	}
	if r.Type != ReservationTypeSoft {
		return fmt.Errorf("%w: only soft reservations can confirm hard (got %s)", ErrInvalidTransition, r.Type)
	}
	r.Type = ReservationTypeHard
	r.ExpiresAt = nil
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// Release frees an active soft or hard reservation (Soft→Released, Hard→Released).
func (r *Reservation) Release() error {
	return r.applyStatus(ReservationStatusReleased)
}

// Consume finalizes a hard reservation (Hard→Consumed on ship/deduct).
func (r *Reservation) Consume() error {
	if r.Type != ReservationTypeHard {
		return fmt.Errorf("%w: only hard reservations can be consumed", ErrInvalidTransition)
	}
	return r.applyStatus(ReservationStatusConsumed)
}

// Expire marks an active soft reservation past TTL (Soft→Expired).
func (r *Reservation) Expire(asOf time.Time) error {
	if r.Type != ReservationTypeSoft {
		return fmt.Errorf("%w: only soft reservations expire", ErrInvalidTransition)
	}
	if !r.IsExpiredAt(asOf) && r.ExpiresAt != nil {
		return fmt.Errorf("%w: not yet past expires_at", ErrInvariant)
	}
	return r.applyStatus(ReservationStatusExpired)
}

// ExtendSoft pushes expires_at forward for an active soft hold.
func (r *Reservation) ExtendSoft(newExpiry time.Time) error {
	if !r.IsActive() {
		return fmt.Errorf("%w: status %s", ErrReservationInactive, r.Status)
	}
	if r.Type != ReservationTypeSoft {
		return fmt.Errorf("%w: only soft reservations can be extended", ErrInvalidTransition)
	}
	if newExpiry.IsZero() {
		return fmt.Errorf("%w: new expiry required", ErrInvalidArgument)
	}
	r.ExpiresAt = &newExpiry
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// TotalQty sums line quantities.
func (r Reservation) TotalQty() int64 {
	var total int64
	for _, l := range r.Lines {
		total += l.Qty
	}
	return total
}
