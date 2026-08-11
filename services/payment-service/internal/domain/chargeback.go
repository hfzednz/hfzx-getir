package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ChargebackStatus tracks dispute lifecycle.
type ChargebackStatus string

const (
	ChargebackOpened   ChargebackStatus = "opened"
	ChargebackWon      ChargebackStatus = "won"
	ChargebackLost     ChargebackStatus = "lost"
	ChargebackCanceled ChargebackStatus = "canceled"
)

// Chargeback records a provider dispute against a captured payment.
type Chargeback struct {
	ID            uuid.UUID
	IntentID      uuid.UUID
	TenantID      uuid.UUID
	AmountMinor   int64
	Currency      string
	Status        ChargebackStatus
	Provider      string
	ProviderRef   string
	ReasonCode    string
	Reason        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks chargeback invariants.
func (c Chargeback) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: chargeback id required", ErrInvalidArgument)
	}
	if c.IntentID == uuid.Nil {
		return fmt.Errorf("%w: intent_id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if _, err := NewMoney(c.AmountMinor, c.Currency); err != nil {
		return err
	}
	return nil
}
