package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DispatchUnitStatus / DispatchStatus is the dispatch queue lifecycle.
type DispatchUnitStatus string

// DispatchStatus is an alias used by app layer.
type DispatchStatus = DispatchUnitStatus

const (
	DispatchStatusQueued     DispatchUnitStatus = "queued"
	DispatchStatusVerified   DispatchUnitStatus = "verified"
	DispatchStatusHandedOff  DispatchUnitStatus = "handed_off"
	DispatchStatusFailedPick DispatchUnitStatus = "failed_pickup"

	// Aliases matching migration enum names.
	DispatchQueued     = DispatchStatusQueued
	DispatchVerified   = DispatchStatusVerified
	DispatchHandedOff  = DispatchStatusHandedOff
	DispatchFailedPick = DispatchStatusFailedPick
)

func (s DispatchUnitStatus) Valid() bool {
	switch s {
	case DispatchStatusQueued, DispatchStatusVerified, DispatchStatusHandedOff, DispatchStatusFailedPick:
		return true
	default:
		return false
	}
}

func (s DispatchUnitStatus) IsTerminal() bool {
	return s == DispatchStatusHandedOff || s == DispatchStatusFailedPick
}

var dispatchTransitions = map[DispatchUnitStatus][]DispatchUnitStatus{
	DispatchStatusQueued: {
		DispatchStatusVerified, DispatchStatusFailedPick, DispatchStatusHandedOff,
	},
	DispatchStatusVerified: {
		DispatchStatusHandedOff, DispatchStatusFailedPick,
	},
}

// DispatchUnit is a packed package awaiting courier handoff.
type DispatchUnit struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	TaskID        uuid.UUID
	PackSessionID *uuid.UUID
	StationID     *uuid.UUID
	LabelID       *uuid.UUID
	PackageCode   string
	TrackingCode  string
	CourierRef    string // opaque dispatch-service / courier assignment
	Status        DispatchUnitStatus
	VerifiedAt    *time.Time
	HandedOffAt   *time.Time
	FailedAt      *time.Time
	Metadata      map[string]any
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks structural invariants.
func (d DispatchUnit) Validate() error {
	if d.ID == uuid.Nil {
		return fmt.Errorf("%w: dispatch unit id required", ErrInvalidArgument)
	}
	if d.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if d.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if d.FulfillmentID == uuid.Nil {
		return fmt.Errorf("%w: fulfillment_id required", ErrInvalidArgument)
	}
	if d.PackageCode == "" && d.TrackingCode == "" {
		return fmt.Errorf("%w: package_code or tracking_code required", ErrInvalidArgument)
	}
	if !d.Status.Valid() {
		return fmt.Errorf("%w: invalid dispatch status %q", ErrInvalidArgument, d.Status)
	}
	return nil
}

// CanTransitionTo reports whether status allows moving to next.
func (d DispatchUnit) CanTransitionTo(next DispatchUnitStatus) bool {
	if !d.Status.Valid() || !next.Valid() {
		return false
	}
	if d.Status == next {
		return true
	}
	if d.Status.IsTerminal() {
		return false
	}
	for _, s := range dispatchTransitions[d.Status] {
		if s == next {
			return true
		}
	}
	return false
}

func (d *DispatchUnit) applyStatus(next DispatchUnitStatus) error {
	if d.Status.IsTerminal() {
		return fmt.Errorf("%w: status %s", ErrAlreadyTerminal, d.Status)
	}
	if !d.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, d.Status, next)
	}
	now := time.Now().UTC()
	d.Status = next
	d.UpdatedAt = now
	switch next {
	case DispatchStatusVerified:
		d.VerifiedAt = &now
	case DispatchStatusHandedOff:
		d.HandedOffAt = &now
	case DispatchStatusFailedPick:
		d.FailedAt = &now
	}
	return nil
}

// VerifyPackage marks package verified in the dispatch queue.
func (d *DispatchUnit) VerifyPackage() error {
	return d.applyStatus(DispatchStatusVerified)
}

// AssignCourier attaches an opaque courier_ref (from dispatch-service).
func (d *DispatchUnit) AssignCourier(courierRef string) error {
	if courierRef == "" {
		return fmt.Errorf("%w: courier_ref required", ErrInvalidArgument)
	}
	if d.Status.IsTerminal() {
		return fmt.Errorf("%w: status %s", ErrAlreadyTerminal, d.Status)
	}
	d.CourierRef = courierRef
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// ConfirmHandoff completes QR handoff to courier.
func (d *DispatchUnit) ConfirmHandoff() error {
	return d.applyStatus(DispatchStatusHandedOff)
}

// FailPickup marks failed courier pickup.
func (d *DispatchUnit) FailPickup() error {
	return d.applyStatus(DispatchStatusFailedPick)
}
