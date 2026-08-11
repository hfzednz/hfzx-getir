package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CashbackStatus is the grant lifecycle.
type CashbackStatus string

const (
	CashbackPending CashbackStatus = "pending"
	CashbackIssued  CashbackStatus = "issued"
	CashbackFailed  CashbackStatus = "failed"
)

// Valid reports whether the cashback status is recognized.
func (s CashbackStatus) Valid() bool {
	switch s {
	case CashbackPending, CashbackIssued, CashbackFailed:
		return true
	default:
		return false
	}
}

// CashbackGrant stores loyalty-side cashback intent; wallet owns the ledger.
type CashbackGrant struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	AccountID      uuid.UUID
	PrincipalID    uuid.UUID
	AmountMinor    int64
	Currency       string
	AccountType    string // cashback | promo
	Status         CashbackStatus
	OrderID        *uuid.UUID
	IdempotencyKey string
	WalletRef      string
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks cashback grant invariants.
func (g CashbackGrant) Validate() error {
	if g.ID == uuid.Nil || g.TenantID == uuid.Nil || g.AccountID == uuid.Nil || g.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: cashback ids required", ErrInvalidArgument)
	}
	if g.AmountMinor <= 0 {
		return fmt.Errorf("%w: amount_minor must be > 0", ErrInvalidArgument)
	}
	if len(g.Currency) != 3 {
		return fmt.Errorf("%w: currency must be ISO-4217", ErrInvalidArgument)
	}
	if !g.Status.Valid() {
		return fmt.Errorf("%w: invalid cashback status", ErrInvalidArgument)
	}
	return nil
}
