package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FulfillmentStatus is the warehouse fulfillment pipeline status.
// Matches ARCHITECTURE state machine (+ cancelled | failed).
type FulfillmentStatus string

const (
	FulfillmentStatusReceived       FulfillmentStatus = "received"
	FulfillmentStatusReserved       FulfillmentStatus = "reserved"
	FulfillmentStatusPickQueued     FulfillmentStatus = "pick_queued"
	FulfillmentStatusPicking        FulfillmentStatus = "picking"
	FulfillmentStatusPicked         FulfillmentStatus = "picked"
	FulfillmentStatusPackQueued     FulfillmentStatus = "pack_queued"
	FulfillmentStatusPacking        FulfillmentStatus = "packing"
	FulfillmentStatusPacked         FulfillmentStatus = "packed"
	FulfillmentStatusDispatchQueued FulfillmentStatus = "dispatch_queued"
	FulfillmentStatusDispatched     FulfillmentStatus = "dispatched"
	FulfillmentStatusCancelled      FulfillmentStatus = "cancelled"
	FulfillmentStatusFailed         FulfillmentStatus = "failed"
)

func (s FulfillmentStatus) Valid() bool {
	switch s {
	case FulfillmentStatusReceived, FulfillmentStatusReserved, FulfillmentStatusPickQueued,
		FulfillmentStatusPicking, FulfillmentStatusPicked, FulfillmentStatusPackQueued,
		FulfillmentStatusPacking, FulfillmentStatusPacked, FulfillmentStatusDispatchQueued,
		FulfillmentStatusDispatched, FulfillmentStatusCancelled, FulfillmentStatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether the status is a terminal outcome.
func (s FulfillmentStatus) IsTerminal() bool {
	return s == FulfillmentStatusDispatched || s == FulfillmentStatusCancelled || s == FulfillmentStatusFailed
}

// PickStrategy selects how pick work is grouped and routed.
type PickStrategy string

const (
	PickStrategySingle   PickStrategy = "single"
	PickStrategyBatch    PickStrategy = "batch"
	PickStrategyWave     PickStrategy = "wave"
	PickStrategyZone     PickStrategy = "zone"
	PickStrategyCluster  PickStrategy = "cluster"
	PickStrategyPriority PickStrategy = "priority"
	PickStrategyExpress  PickStrategy = "express"
)

func (s PickStrategy) Valid() bool {
	switch s {
	case PickStrategySingle, PickStrategyBatch, PickStrategyWave, PickStrategyZone,
		PickStrategyCluster, PickStrategyPriority, PickStrategyExpress:
		return true
	default:
		return false
	}
}

// Happy-path transitions from ARCHITECTURE.md; cancel/fail from non-terminal.
var fulfillmentTransitions = map[FulfillmentStatus][]FulfillmentStatus{
	FulfillmentStatusReceived: {
		FulfillmentStatusReserved, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusReserved: {
		FulfillmentStatusPickQueued, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPickQueued: {
		FulfillmentStatusPicking, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPicking: {
		FulfillmentStatusPicked, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPicked: {
		FulfillmentStatusPackQueued, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPackQueued: {
		FulfillmentStatusPacking, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPacking: {
		FulfillmentStatusPacked, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusPacked: {
		FulfillmentStatusDispatchQueued, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
	FulfillmentStatusDispatchQueued: {
		FulfillmentStatusDispatched, FulfillmentStatusCancelled, FulfillmentStatusFailed,
	},
}

// FulfillmentOrder is a warehouse projection of an opaque order.
// ExternalOrderID / ReservationID are opaque — no order aggregate or stock ledger here.
type FulfillmentOrder struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	WarehouseID     uuid.UUID
	ExternalOrderID string // opaque order-service id (wire form)
	OrderID         uuid.UUID // opaque UUID form when provided
	ReservationID   *uuid.UUID // opaque inventory reservation id
	Status          FulfillmentStatus
	Priority        int
	Strategy        PickStrategy
	VIP             bool
	Express         bool
	SLADeadline     *time.Time
	CourierRef      string // opaque courier/dispatch ref after handoff
	Lines           []FulfillmentLine
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CancelledAt     *time.Time
	FailedAt        *time.Time
	CompletedAt     *time.Time
}

// FulfillmentLine is a pick/pack line with opaque variant/location/barcode.
type FulfillmentLine struct {
	ID              uuid.UUID
	FulfillmentID   uuid.UUID
	VariantID       uuid.UUID // opaque catalog variant
	SKUCode         string
	Barcode         string // expected scan target (opaque)
	BarcodeExpected string // alias preferred by ValidatePickScan path
	LocationCode    string // opaque inventory location code
	QtyOrdered      int64
	Qty             int // alias for QtyOrdered when using int domain APIs
	QtyPicked       int64
	QtyPacked       int64
	ExpiryRequired  bool
	Sequence        int
	SortOrder       int
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ExpectedBarcode returns the scan target (Barcode or BarcodeExpected).
func (l FulfillmentLine) ExpectedBarcode() string {
	if l.BarcodeExpected != "" {
		return l.BarcodeExpected
	}
	return l.Barcode
}

// OrderedQty returns the ordered quantity (QtyOrdered preferred).
func (l FulfillmentLine) OrderedQty() int64 {
	if l.QtyOrdered > 0 {
		return l.QtyOrdered
	}
	return int64(l.Qty)
}

// Validate checks structural invariants.
func (f FulfillmentOrder) Validate() error {
	if f.ID == uuid.Nil {
		return fmt.Errorf("%w: fulfillment id required", ErrInvalidArgument)
	}
	if f.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if f.ExternalOrderID == "" && f.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if f.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !f.Status.Valid() {
		return fmt.Errorf("%w: invalid fulfillment status %q", ErrInvalidArgument, f.Status)
	}
	if f.Strategy != "" && !f.Strategy.Valid() {
		return fmt.Errorf("%w: invalid pick strategy %q", ErrInvalidArgument, f.Strategy)
	}
	for i, line := range f.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks a fulfillment line.
func (l FulfillmentLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: line id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.OrderedQty() <= 0 {
		return fmt.Errorf("%w: qty must be positive", ErrInvalidArgument)
	}
	if l.QtyPicked < 0 || l.QtyPicked > l.OrderedQty() {
		return fmt.Errorf("%w: qty_picked out of range", ErrInvariant)
	}
	if l.QtyPacked < 0 || l.QtyPacked > l.QtyPicked {
		return fmt.Errorf("%w: qty_packed out of range", ErrInvariant)
	}
	return nil
}

// QtyRemaining returns units still to pick on the line.
func (l FulfillmentLine) QtyRemaining() int64 {
	return l.OrderedQty() - l.QtyPicked
}

// IsFullyPicked reports whether all units are picked.
func (l FulfillmentLine) IsFullyPicked() bool {
	return l.QtyPicked >= l.OrderedQty()
}

// CanTransitionTo reports whether status allows moving to next.
func (f FulfillmentOrder) CanTransitionTo(next FulfillmentStatus) bool {
	if !f.Status.Valid() || !next.Valid() {
		return false
	}
	if f.Status == next {
		return true
	}
	if f.Status.IsTerminal() {
		return false
	}
	for _, s := range fulfillmentTransitions[f.Status] {
		if s == next {
			return true
		}
	}
	return false
}

func (f *FulfillmentOrder) applyStatus(next FulfillmentStatus) error {
	if f.Status.IsTerminal() {
		return fmt.Errorf("%w: status %s", ErrAlreadyTerminal, f.Status)
	}
	if !f.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, f.Status, next)
	}
	now := time.Now().UTC()
	f.Status = next
	f.UpdatedAt = now
	switch next {
	case FulfillmentStatusCancelled:
		f.CancelledAt = &now
	case FulfillmentStatusFailed:
		f.FailedAt = &now
	case FulfillmentStatusDispatched:
		f.CompletedAt = &now
	}
	return nil
}

// MarkReserved moves received → reserved after inventory Soft/Hard via port.
func (f *FulfillmentOrder) MarkReserved(reservationID uuid.UUID) error {
	if reservationID == uuid.Nil {
		return fmt.Errorf("%w: reservation_id required", ErrInvalidArgument)
	}
	if err := f.applyStatus(FulfillmentStatusReserved); err != nil {
		return err
	}
	f.ReservationID = &reservationID
	return nil
}

// QueuePick moves reserved → pick_queued (CreatePickTask).
func (f *FulfillmentOrder) QueuePick() error {
	return f.applyStatus(FulfillmentStatusPickQueued)
}

// StartPicking moves pick_queued → picking (AssignPicker).
func (f *FulfillmentOrder) StartPicking() error {
	return f.applyStatus(FulfillmentStatusPicking)
}

// CompletePicking moves picking → picked (ScanComplete).
func (f *FulfillmentOrder) CompletePicking() error {
	return f.applyStatus(FulfillmentStatusPicked)
}

// QueuePack moves picked → pack_queued (CreatePackTask).
func (f *FulfillmentOrder) QueuePack() error {
	return f.applyStatus(FulfillmentStatusPackQueued)
}

// StartPacking moves pack_queued → packing (AssignPacker).
func (f *FulfillmentOrder) StartPacking() error {
	return f.applyStatus(FulfillmentStatusPacking)
}

// CompletePacking moves packing → packed (WeightLabelOK).
func (f *FulfillmentOrder) CompletePacking() error {
	return f.applyStatus(FulfillmentStatusPacked)
}

// QueueDispatch moves packed → dispatch_queued.
func (f *FulfillmentOrder) QueueDispatch() error {
	return f.applyStatus(FulfillmentStatusDispatchQueued)
}

// CompleteDispatch moves dispatch_queued → dispatched (CourierHandoff).
func (f *FulfillmentOrder) CompleteDispatch() error {
	return f.applyStatus(FulfillmentStatusDispatched)
}

// Cancel moves to cancelled from any non-terminal status.
func (f *FulfillmentOrder) Cancel() error {
	return f.applyStatus(FulfillmentStatusCancelled)
}

// Fail moves to failed from any non-terminal status.
func (f *FulfillmentOrder) Fail() error {
	return f.applyStatus(FulfillmentStatusFailed)
}

// AllLinesPicked reports whether every line is fully picked.
func (f FulfillmentOrder) AllLinesPicked() bool {
	if len(f.Lines) == 0 {
		return false
	}
	for _, l := range f.Lines {
		if !l.IsFullyPicked() {
			return false
		}
	}
	return true
}
