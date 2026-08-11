package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProviderRoute maps a method type to ordered PSP providers (failover).
type ProviderRoute struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	MethodType   PaymentMethodType
	Currency     string
	Providers    []string // ordered preference; first = primary
	Active       bool
	Priority     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate checks route invariants.
func (r ProviderRoute) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: route id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if len(r.Providers) == 0 {
		return fmt.Errorf("%w: at least one provider required", ErrInvalidArgument)
	}
	return nil
}
