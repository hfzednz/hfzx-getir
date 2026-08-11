package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PaymentMethod stores tokenized payment credentials only (never PAN).
type PaymentMethod struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	PrincipalID  uuid.UUID
	MethodType   PaymentMethodType
	Token        string // opaque PSP/HSM token
	Last4        string
	Brand        string
	ExpMonth     int
	ExpYear      int
	Provider     string
	Active       bool
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks payment method invariants (token only, no PAN).
func (m PaymentMethod) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: method id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if m.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if m.Token == "" {
		return fmt.Errorf("%w: token required (PAN never stored)", ErrInvalidArgument)
	}
	return nil
}
