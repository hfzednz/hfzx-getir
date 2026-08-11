package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefundStatus is the refund lifecycle.
type RefundStatus string

const (
	RefundRequested RefundStatus = "requested"
	RefundCompleted RefundStatus = "completed"
	RefundFailed    RefundStatus = "failed"
)

// Refund is a (partial or full) refund against a captured intent.
type Refund struct {
	ID               uuid.UUID
	IntentID         uuid.UUID
	TenantID         uuid.UUID
	AmountMinor      int64
	Currency         string
	Status           RefundStatus
	Provider         string
	ProviderRef      string
	Reason           string
	IdempotencyKey   string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}

// Validate checks refund invariants.
func (r Refund) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: refund id required", ErrInvalidArgument)
	}
	if r.IntentID == uuid.Nil {
		return fmt.Errorf("%w: intent_id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if r.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	if _, err := NewMoney(r.AmountMinor, r.Currency); err != nil {
		return err
	}
	if r.AmountMinor == 0 {
		return fmt.Errorf("%w: refund amount must be > 0", ErrInvalidArgument)
	}
	return nil
}
