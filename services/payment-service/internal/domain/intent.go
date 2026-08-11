package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IntentStatus is the payment intent lifecycle.
type IntentStatus string

const (
	IntentInitiated  IntentStatus = "initiated"
	IntentAuthorized IntentStatus = "authorized"
	IntentCaptured   IntentStatus = "captured"
	IntentVoided     IntentStatus = "voided"
	IntentFailed     IntentStatus = "failed"
	IntentRefunded   IntentStatus = "refunded" // full refund after capture
)

// Valid reports whether the status is recognized.
func (s IntentStatus) Valid() bool {
	switch s {
	case IntentInitiated, IntentAuthorized, IntentCaptured, IntentVoided, IntentFailed, IntentRefunded:
		return true
	default:
		return false
	}
}

// PaymentMethodType classifies how the intent is funded.
type PaymentMethodType string

const (
	MethodCard   PaymentMethodType = "card"
	MethodWallet PaymentMethodType = "wallet"
	MethodApple  PaymentMethodType = "apple_pay"
	MethodGoogle PaymentMethodType = "google_pay"
)

// PaymentIntent is the payment orchestration aggregate.
// order_id is an opaque cross-service reference (no order engine here).
type PaymentIntent struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	PrincipalID        uuid.UUID
	OrderID            string // opaque
	Status             IntentStatus
	AmountMinor        int64
	CapturedMinor      int64
	RefundedMinor      int64
	Currency           string
	MethodType         PaymentMethodType
	PaymentMethodID    *uuid.UUID
	Provider           string
	ProviderIntentRef  string
	IdempotencyKey     string
	FraudScore         int
	FraudDecision      string
	FailureReason      string
	Metadata           map[string]any
	Version            int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	AuthorizedAt       *time.Time
	CapturedAt         *time.Time
	VoidedAt           *time.Time
	FailedAt           *time.Time
}

// RemainingCapturable returns authorized - captured (for authorized intents).
func (i PaymentIntent) RemainingCapturable() int64 {
	if i.Status != IntentAuthorized && i.Status != IntentCaptured {
		return 0
	}
	rem := i.AmountMinor - i.CapturedMinor
	if rem < 0 {
		return 0
	}
	return rem
}

// RemainingRefundable returns captured - refunded.
func (i PaymentIntent) RemainingRefundable() int64 {
	rem := i.CapturedMinor - i.RefundedMinor
	if rem < 0 {
		return 0
	}
	return rem
}

// Validate checks structural invariants.
func (i PaymentIntent) Validate() error {
	if i.ID == uuid.Nil {
		return fmt.Errorf("%w: intent id required", ErrInvalidArgument)
	}
	if i.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if i.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if !i.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, i.Status)
	}
	if i.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	m, err := NewMoney(i.AmountMinor, i.Currency)
	if err != nil {
		return err
	}
	_ = m
	if i.CapturedMinor < 0 || i.RefundedMinor < 0 {
		return fmt.Errorf("%w: captured/refunded must be non-negative", ErrNegativeMoney)
	}
	if i.CapturedMinor > i.AmountMinor {
		return fmt.Errorf("%w: captured exceeds amount", ErrInvariant)
	}
	if i.RefundedMinor > i.CapturedMinor {
		return fmt.Errorf("%w: refunded exceeds captured", ErrInvariant)
	}
	if i.Version < 1 {
		return fmt.Errorf("%w: version must be >= 1", ErrInvalidArgument)
	}
	return nil
}

// ValidateTransition checks intent status machine transitions.
func ValidateTransition(from, to IntentStatus) error {
	if from == to {
		return nil
	}
	ok := false
	switch from {
	case IntentInitiated:
		ok = to == IntentAuthorized || to == IntentFailed
	case IntentAuthorized:
		ok = to == IntentCaptured || to == IntentVoided || to == IntentFailed
	case IntentCaptured:
		ok = to == IntentRefunded // full refund only; partial keeps captured
	case IntentVoided, IntentFailed, IntentRefunded:
		ok = false
	}
	if !ok {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}
