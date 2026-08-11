package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AttemptKind classifies a PSP attempt.
type AttemptKind string

const (
	AttemptAuthorize AttemptKind = "authorize"
	AttemptCapture   AttemptKind = "capture"
	AttemptVoid      AttemptKind = "void"
	AttemptRefund    AttemptKind = "refund"
)

// AttemptStatus is the outcome of a PSP call.
type AttemptStatus string

const (
	AttemptPending AttemptStatus = "pending"
	AttemptSuccess AttemptStatus = "success"
	AttemptFailed  AttemptStatus = "failed"
)

// PaymentAttempt records a single PSP (or wallet) operation attempt.
type PaymentAttempt struct {
	ID               uuid.UUID
	IntentID         uuid.UUID
	TenantID         uuid.UUID
	Kind             AttemptKind
	Status           AttemptStatus
	Provider         string
	ProviderRef      string
	AmountMinor      int64
	Currency         string
	ErrorCode        string
	ErrorMessage     string
	IdempotencyKey   string
	IsFailover       bool
	CreatedAt        time.Time
}

// Validate checks attempt invariants.
func (a PaymentAttempt) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: attempt id required", ErrInvalidArgument)
	}
	if a.IntentID == uuid.Nil {
		return fmt.Errorf("%w: intent_id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if a.AmountMinor < 0 {
		return fmt.Errorf("%w: amount", ErrNegativeMoney)
	}
	return nil
}
