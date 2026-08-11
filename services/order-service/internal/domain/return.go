package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReturnStatus is the OMS return request lifecycle.
type ReturnStatus string

const (
	ReturnStatusRequested ReturnStatus = "requested"
	ReturnStatusApproved  ReturnStatus = "approved"
	ReturnStatusRejected  ReturnStatus = "rejected"
	ReturnStatusInTransit ReturnStatus = "in_transit"
	ReturnStatusReceived  ReturnStatus = "received"
	ReturnStatusCompleted ReturnStatus = "completed"
	ReturnStatusCancelled ReturnStatus = "cancelled"
)

// Valid reports whether the return status is recognized.
func (s ReturnStatus) Valid() bool {
	switch s {
	case ReturnStatusRequested, ReturnStatusApproved, ReturnStatusRejected,
		ReturnStatusInTransit, ReturnStatusReceived, ReturnStatusCompleted,
		ReturnStatusCancelled:
		return true
	default:
		return false
	}
}

var returnTransitions = map[ReturnStatus][]ReturnStatus{
	ReturnStatusRequested: {
		ReturnStatusApproved, ReturnStatusRejected, ReturnStatusCancelled,
	},
	ReturnStatusApproved: {
		ReturnStatusInTransit, ReturnStatusReceived, ReturnStatusCancelled,
	},
	ReturnStatusInTransit: {
		ReturnStatusReceived, ReturnStatusCancelled,
	},
	ReturnStatusReceived: {
		ReturnStatusCompleted,
	},
	ReturnStatusRejected:  {},
	ReturnStatusCompleted: {},
	ReturnStatusCancelled: {},
}

// CanTransitionTo reports whether from → to is allowed for returns.
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

// ReturnDisposition is the intended inventory disposition hint (no stock write here).
type ReturnDisposition string

const (
	ReturnDispositionRestock    ReturnDisposition = "restock"
	ReturnDispositionQuarantine ReturnDisposition = "quarantine"
	ReturnDispositionWaste      ReturnDisposition = "waste"
	ReturnDispositionResell     ReturnDisposition = "resell"
	ReturnDispositionPending    ReturnDisposition = "pending"
)

// Valid reports whether the disposition is recognized.
func (d ReturnDisposition) Valid() bool {
	switch d {
	case ReturnDispositionRestock, ReturnDispositionQuarantine, ReturnDispositionWaste,
		ReturnDispositionResell, ReturnDispositionPending:
		return true
	default:
		return false
	}
}

// Return is an OMS return request header (stock write-back via inventory port/events).
type Return struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	TenantID    uuid.UUID
	Status      ReturnStatus
	Disposition ReturnDisposition
	Reason      string
	Notes       string
	ActorID     *uuid.UUID
	RefundID    *uuid.UUID
	Lines       []ReturnLine
	Metadata    map[string]any
	RequestedAt time.Time
	DecidedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ReturnLine references an original order line being returned.
type ReturnLine struct {
	ID          uuid.UUID
	ReturnID    uuid.UUID
	OrderLineID uuid.UUID
	VariantID   uuid.UUID
	Qty         int
	Disposition ReturnDisposition
	Reason      string
	Metadata    map[string]any
	CreatedAt   time.Time
}

// Validate checks return header invariants.
func (r Return) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: return id required", ErrInvalidArgument)
	}
	if r.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid return status %q", ErrInvalidArgument, r.Status)
	}
	if !r.Disposition.Valid() {
		return fmt.Errorf("%w: invalid disposition %q", ErrInvalidArgument, r.Disposition)
	}
	if len(r.Lines) == 0 {
		return fmt.Errorf("%w: return requires at least one line", ErrInvalidArgument)
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
	if l.OrderLineID == uuid.Nil {
		return fmt.Errorf("%w: order_line_id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	if !l.Disposition.Valid() {
		return fmt.Errorf("%w: invalid disposition %q", ErrInvalidArgument, l.Disposition)
	}
	return nil
}

// TransitionTo applies a validated return status change.
func (r *Return) TransitionTo(next ReturnStatus) error {
	if !r.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: return %s → %s", ErrInvalidTransition, r.Status, next)
	}
	if r.Status == next {
		return nil
	}
	now := time.Now().UTC()
	r.Status = next
	r.UpdatedAt = now
	switch next {
	case ReturnStatusApproved, ReturnStatusRejected:
		r.DecidedAt = &now
	case ReturnStatusCompleted:
		r.CompletedAt = &now
	}
	return nil
}

// CanRequestReturn reports whether the order status allows a return request.
func CanRequestReturn(orderStatus OrderStatus) bool {
	switch orderStatus {
	case OrderStatusDelivered, OrderStatusCompleted:
		return true
	default:
		return false
	}
}

// AssertReturnAllowed returns ErrReturnNotAllowed when returns are blocked.
func AssertReturnAllowed(orderStatus OrderStatus) error {
	if !CanRequestReturn(orderStatus) {
		return fmt.Errorf("%w: status %s", ErrReturnNotAllowed, orderStatus)
	}
	return nil
}
