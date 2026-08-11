package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReturnSource identifies where the returned stock originated.
type ReturnSource string

const (
	ReturnSourceCustomer  ReturnSource = "customer"
	ReturnSourceCourier   ReturnSource = "courier"
	ReturnSourceWarehouse ReturnSource = "warehouse"
	ReturnSourceSupplier  ReturnSource = "supplier"
)

func (s ReturnSource) Valid() bool {
	switch s {
	case ReturnSourceCustomer, ReturnSourceCourier, ReturnSourceWarehouse, ReturnSourceSupplier:
		return true
	default:
		return false
	}
}

// ReturnDisposition decides how returned qty re-enters (or leaves) inventory.
type ReturnDisposition string

const (
	ReturnDispositionRestock    ReturnDisposition = "restock"
	ReturnDispositionQuarantine ReturnDisposition = "quarantine"
	ReturnDispositionWaste      ReturnDisposition = "waste"
)

func (d ReturnDisposition) Valid() bool {
	switch d {
	case ReturnDispositionRestock, ReturnDispositionQuarantine, ReturnDispositionWaste:
		return true
	default:
		return false
	}
}

// ReturnStatus is the lifecycle of an inventory return header.
type ReturnStatus string

const (
	ReturnStatusDraft     ReturnStatus = "draft"
	ReturnStatusReceived  ReturnStatus = "received"
	ReturnStatusDisposed  ReturnStatus = "disposed"
	ReturnStatusCancelled ReturnStatus = "cancelled"
)

func (s ReturnStatus) Valid() bool {
	switch s {
	case ReturnStatusDraft, ReturnStatusReceived, ReturnStatusDisposed, ReturnStatusCancelled:
		return true
	default:
		return false
	}
}

var returnTransitions = map[ReturnStatus][]ReturnStatus{
	ReturnStatusDraft: {
		ReturnStatusReceived, ReturnStatusCancelled,
	},
	ReturnStatusReceived: {
		ReturnStatusDisposed, ReturnStatusCancelled,
	},
	ReturnStatusDisposed:  {},
	ReturnStatusCancelled: {},
}

// CanTransitionTo reports whether from → to is allowed.
func (s ReturnStatus) CanTransitionTo(to ReturnStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	if s == to {
		return true
	}
	for _, next := range returnTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// InventoryReturn is a return-to-stock header (opaque external_ref only).
type InventoryReturn struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Source      ReturnSource
	Disposition ReturnDisposition
	Status      ReturnStatus
	ExternalRef string
	ActorID     *uuid.UUID
	Reason      string
	Lines       []ReturnLine
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ReceivedAt  *time.Time
	DisposedAt  *time.Time
}

// ReturnLine is a returned variant quantity.
type ReturnLine struct {
	ID             uuid.UUID
	ReturnID       uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	LotID          *uuid.UUID
	LocationID     *uuid.UUID
	Qty            int64
	Disposition    *ReturnDisposition
	ConditionNotes string
	Metadata       map[string]any
	CreatedAt      time.Time
}

// Validate checks structural invariants.
func (r InventoryReturn) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: return id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if r.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !r.Source.Valid() {
		return fmt.Errorf("%w: invalid return source %q", ErrInvalidArgument, r.Source)
	}
	if !r.Disposition.Valid() {
		return fmt.Errorf("%w: invalid disposition %q", ErrInvalidArgument, r.Disposition)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid return status %q", ErrInvalidArgument, r.Status)
	}
	for i, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks a return line.
func (l ReturnLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: return line id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: return qty", ErrNegativeQuantity)
	}
	if l.Disposition != nil && !l.Disposition.Valid() {
		return fmt.Errorf("%w: invalid line disposition %q", ErrInvalidArgument, *l.Disposition)
	}
	return nil
}

// EffectiveDisposition returns line override or header disposition.
func (l ReturnLine) EffectiveDisposition(header ReturnDisposition) ReturnDisposition {
	if l.Disposition != nil {
		return *l.Disposition
	}
	return header
}

// TransitionTo moves the return along the allowed state machine.
func (r *InventoryReturn) TransitionTo(next ReturnStatus) error {
	if !r.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, r.Status, next)
	}
	now := time.Now().UTC()
	r.Status = next
	r.UpdatedAt = now
	switch next {
	case ReturnStatusReceived:
		r.ReceivedAt = &now
	case ReturnStatusDisposed:
		r.DisposedAt = &now
	}
	return nil
}

// TotalQty sums line quantities.
func (r InventoryReturn) TotalQty() int64 {
	var total int64
	for _, l := range r.Lines {
		total += l.Qty
	}
	return total
}
