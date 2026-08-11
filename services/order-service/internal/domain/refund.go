package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefundMethod is how funds are returned (executed by payment-service port).
type RefundMethod string

const (
	RefundMethodWallet      RefundMethod = "wallet"
	RefundMethodCard        RefundMethod = "card"
	RefundMethodStoreCredit RefundMethod = "store_credit"
)

// Valid reports whether the refund method is recognized.
func (m RefundMethod) Valid() bool {
	switch m {
	case RefundMethodWallet, RefundMethodCard, RefundMethodStoreCredit:
		return true
	default:
		return false
	}
}

// RefundStatus is the OMS refund request lifecycle.
type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "pending"
	RefundStatusAuthorized RefundStatus = "authorized"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusSucceeded  RefundStatus = "succeeded"
	RefundStatusFailed     RefundStatus = "failed"
	RefundStatusCancelled  RefundStatus = "cancelled"
)

// Valid reports whether the refund status is recognized.
func (s RefundStatus) Valid() bool {
	switch s {
	case RefundStatusPending, RefundStatusAuthorized, RefundStatusProcessing,
		RefundStatusSucceeded, RefundStatusFailed, RefundStatusCancelled:
		return true
	default:
		return false
	}
}

var refundTransitions = map[RefundStatus][]RefundStatus{
	RefundStatusPending: {
		RefundStatusAuthorized, RefundStatusProcessing, RefundStatusCancelled, RefundStatusFailed,
	},
	RefundStatusAuthorized: {
		RefundStatusProcessing, RefundStatusCancelled, RefundStatusFailed,
	},
	RefundStatusProcessing: {
		RefundStatusSucceeded, RefundStatusFailed,
	},
	RefundStatusSucceeded: {},
	RefundStatusFailed: {
		RefundStatusPending, // retry
		RefundStatusCancelled,
	},
	RefundStatusCancelled: {},
}

// CanTransitionTo reports whether from → to is allowed for refunds.
func (s RefundStatus) CanTransitionTo(to RefundStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	if s == to {
		return true
	}
	for _, next := range refundTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// Refund is an OMS refund request. PSP/wallet execution is via opaque payment_refund_ref.
type Refund struct {
	ID               uuid.UUID
	OrderID          uuid.UUID
	TenantID         uuid.UUID
	ReturnID         *uuid.UUID
	AmountMinor      int64
	Currency         string
	Method           RefundMethod
	Status           RefundStatus
	Reason           string
	PaymentRefundRef string // opaque payment-service refund id
	ActorID          *uuid.UUID
	Metadata         map[string]any
	RequestedAt      time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks refund invariants (integer minor units only).
func (r Refund) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: refund id required", ErrInvalidArgument)
	}
	if r.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if r.AmountMinor <= 0 {
		return fmt.Errorf("%w: amount_minor must be > 0", ErrInvalidArgument)
	}
	if _, err := NewMoney(r.AmountMinor, r.Currency); err != nil {
		return err
	}
	if !r.Method.Valid() {
		return fmt.Errorf("%w: invalid refund method %q", ErrInvalidArgument, r.Method)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid refund status %q", ErrInvalidArgument, r.Status)
	}
	return nil
}

// Amount returns the refund amount as Money.
func (r Refund) Amount() (Money, error) {
	return NewMoney(r.AmountMinor, r.Currency)
}

// TransitionTo applies a validated refund status change.
func (r *Refund) TransitionTo(next RefundStatus) error {
	if !r.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: refund %s → %s", ErrInvalidTransition, r.Status, next)
	}
	if r.Status == next {
		return nil
	}
	now := time.Now().UTC()
	r.Status = next
	r.UpdatedAt = now
	if next == RefundStatusSucceeded {
		r.CompletedAt = &now
	}
	return nil
}

// CanRequestRefund reports whether the order status allows a refund request.
func CanRequestRefund(orderStatus OrderStatus) bool {
	switch orderStatus {
	case OrderStatusDelivered, OrderStatusCompleted, OrderStatusRefundPending,
		OrderStatusCancelled:
		// Cancelled may still need post-cancel refund completion.
		return true
	default:
		return false
	}
}

// AssertRefundAllowed returns ErrRefundNotAllowed when refunds are blocked.
func AssertRefundAllowed(orderStatus OrderStatus) error {
	if !CanRequestRefund(orderStatus) {
		return fmt.Errorf("%w: status %s", ErrRefundNotAllowed, orderStatus)
	}
	return nil
}
