package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrderStatus is the canonical OMS lifecycle status (ARCHITECTURE.md).
type OrderStatus string

const (
	OrderStatusDraft                OrderStatus = "draft"
	OrderStatusPendingPayment       OrderStatus = "pending_payment"
	OrderStatusPaymentProcessing    OrderStatus = "payment_processing"
	OrderStatusPaymentFailed        OrderStatus = "payment_failed"
	OrderStatusInventoryReservation OrderStatus = "inventory_reservation"
	OrderStatusInventoryFailed      OrderStatus = "inventory_failed"
	OrderStatusWarehouseAssigned    OrderStatus = "warehouse_assigned"
	OrderStatusPicking              OrderStatus = "picking"
	OrderStatusPacking              OrderStatus = "packing"
	OrderStatusReadyForDispatch     OrderStatus = "ready_for_dispatch"
	OrderStatusCourierAssigned      OrderStatus = "courier_assigned"
	OrderStatusOutForDelivery       OrderStatus = "out_for_delivery"
	OrderStatusDelivered            OrderStatus = "delivered"
	OrderStatusCompleted            OrderStatus = "completed"
	OrderStatusCancelled            OrderStatus = "cancelled"
	OrderStatusRefundPending        OrderStatus = "refund_pending"
	OrderStatusRefunded             OrderStatus = "refunded"
	OrderStatusFailed               OrderStatus = "failed"
	OrderStatusArchived             OrderStatus = "archived"
)

// Valid reports whether the status is a known OMS status.
func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusDraft, OrderStatusPendingPayment, OrderStatusPaymentProcessing,
		OrderStatusPaymentFailed, OrderStatusInventoryReservation, OrderStatusInventoryFailed,
		OrderStatusWarehouseAssigned, OrderStatusPicking, OrderStatusPacking,
		OrderStatusReadyForDispatch, OrderStatusCourierAssigned, OrderStatusOutForDelivery,
		OrderStatusDelivered, OrderStatusCompleted, OrderStatusCancelled,
		OrderStatusRefundPending, OrderStatusRefunded, OrderStatusFailed, OrderStatusArchived:
		return true
	default:
		return false
	}
}

// IsTerminal reports whether no further happy-path progression is expected
// (archived is the only fully sealed terminal; cancelled/refunded/failed may archive).
func (s OrderStatus) IsTerminal() bool {
	switch s {
	case OrderStatusCancelled, OrderStatusRefunded, OrderStatusFailed, OrderStatusArchived,
		OrderStatusCompleted, OrderStatusPaymentFailed, OrderStatusInventoryFailed:
		return true
	default:
		return false
	}
}

// Strict transition table. Illegal jumps → ErrInvalidTransition.
// draft→pending_payment allowed; draft→delivered NOT allowed.
var orderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusDraft: {
		OrderStatusPendingPayment,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusPendingPayment: {
		OrderStatusPaymentProcessing,
		OrderStatusInventoryReservation,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusPaymentProcessing: {
		OrderStatusPaymentFailed,
		OrderStatusInventoryReservation,
		OrderStatusWarehouseAssigned,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusPaymentFailed: {
		OrderStatusPendingPayment,
		OrderStatusCancelled,
		OrderStatusFailed,
		OrderStatusArchived,
	},
	OrderStatusInventoryReservation: {
		OrderStatusInventoryFailed,
		OrderStatusPaymentProcessing,
		OrderStatusWarehouseAssigned,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusInventoryFailed: {
		OrderStatusPendingPayment,
		OrderStatusCancelled,
		OrderStatusFailed,
		OrderStatusArchived,
	},
	OrderStatusWarehouseAssigned: {
		OrderStatusPicking,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusPicking: {
		OrderStatusPacking,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusPacking: {
		OrderStatusReadyForDispatch,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusReadyForDispatch: {
		OrderStatusCourierAssigned,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusCourierAssigned: {
		OrderStatusOutForDelivery,
		OrderStatusCancelled,
		OrderStatusFailed,
	},
	OrderStatusOutForDelivery: {
		OrderStatusDelivered,
		OrderStatusFailed,
	},
	OrderStatusDelivered: {
		OrderStatusCompleted,
		OrderStatusRefundPending,
	},
	OrderStatusCompleted: {
		OrderStatusRefundPending,
		OrderStatusArchived,
	},
	OrderStatusRefundPending: {
		OrderStatusRefunded,
		OrderStatusFailed,
		OrderStatusCompleted, // refund abandoned / partial path back
	},
	OrderStatusRefunded: {
		OrderStatusArchived,
	},
	OrderStatusCancelled: {
		OrderStatusArchived,
	},
	OrderStatusFailed: {
		OrderStatusArchived,
	},
	OrderStatusArchived: {},
}

// CanTransitionTo reports whether from → to is a legal OMS transition.
func (s OrderStatus) CanTransitionTo(to OrderStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	if s == to {
		return true
	}
	for _, next := range orderTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// ValidateTransition returns ErrInvalidTransition when from → to is illegal.
func ValidateTransition(from, to OrderStatus) error {
	if !from.Valid() {
		return fmt.Errorf("%w: unknown from status %q", ErrInvalidArgument, from)
	}
	if !to.Valid() {
		return fmt.Errorf("%w: unknown to status %q", ErrInvalidArgument, to)
	}
	if from == to {
		return nil
	}
	if !from.CanTransitionTo(to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}

// Order is the OMS aggregate root. Opaque refs only — no ledger/PSP/WH/courier logic.
type Order struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	CustomerPrincipalID uuid.UUID
	Status              OrderStatus
	Type                OrderType
	Currency            string
	SubtotalMinor       int64
	DiscountMinor       int64
	TaxMinor            int64
	ShippingMinor       int64
	TipMinor            int64
	TotalMinor          int64
	AddressSnapshot     map[string]any
	Notes               string
	Gift                map[string]any
	Priority            int
	WarehouseIDs        []uuid.UUID
	Version             int64
	IdempotencyKey      string
	ScheduledAt         *time.Time
	PlacedAt            *time.Time
	PaymentIntentRef    string // opaque payment-service ref
	ReservationRef      string // opaque inventory-service ref
	CourierRef          string // opaque dispatch-service ref
	ParentOrderID       *uuid.UUID
	Lines               []OrderLine
	Metadata            map[string]any
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CancelledAt         *time.Time
	CompletedAt         *time.Time
	ArchivedAt          *time.Time
}

// Validate checks aggregate structural invariants.
func (o Order) Validate() error {
	if o.ID == uuid.Nil {
		return fmt.Errorf("%w: order id required", ErrInvalidArgument)
	}
	if o.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if o.CustomerPrincipalID == uuid.Nil {
		return fmt.Errorf("%w: customer_principal_id required", ErrInvalidArgument)
	}
	if !o.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, o.Status)
	}
	if err := o.Type.Validate(); err != nil {
		return err
	}
	if _, err := NewMoney(o.TotalMinor, o.Currency); err != nil {
		return fmt.Errorf("total: %w", err)
	}
	if o.SubtotalMinor < 0 || o.DiscountMinor < 0 || o.TaxMinor < 0 ||
		o.ShippingMinor < 0 || o.TipMinor < 0 {
		return fmt.Errorf("%w: money totals", ErrNegativeMoney)
	}
	if o.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	if o.Version < 1 {
		return fmt.Errorf("%w: version must be >= 1", ErrInvalidArgument)
	}
	for i, line := range o.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// CanTransitionTo reports whether this order may move to next.
func (o Order) CanTransitionTo(next OrderStatus) bool {
	return o.Status.CanTransitionTo(next)
}

// TransitionTo applies a validated status change and bumps version.
func (o *Order) TransitionTo(next OrderStatus) error {
	if err := ValidateTransition(o.Status, next); err != nil {
		return err
	}
	if o.Status == next {
		return nil
	}
	now := time.Now().UTC()
	o.Status = next
	o.UpdatedAt = now
	o.Version++
	switch next {
	case OrderStatusPendingPayment:
		if o.PlacedAt == nil {
			o.PlacedAt = &now
		}
	case OrderStatusCancelled:
		o.CancelledAt = &now
	case OrderStatusCompleted:
		o.CompletedAt = &now
	case OrderStatusArchived:
		o.ArchivedAt = &now
	}
	return nil
}

// Total returns the order total as Money.
func (o Order) Total() (Money, error) {
	return NewMoney(o.TotalMinor, o.Currency)
}
