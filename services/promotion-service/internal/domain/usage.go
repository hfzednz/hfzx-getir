package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UsageScope identifies the counter dimension.
type UsageScope string

const (
	UsageGlobal UsageScope = "global"
	UsageUser   UsageScope = "user"
	UsageOrder  UsageScope = "order"
	UsageDevice UsageScope = "device"
)

// UsageCounter tracks redemptions for limit enforcement.
type UsageCounter struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PromotionID uuid.UUID
	Scope       UsageScope
	ScopeKey    string // principal/order/device/global
	Count       int
	UpdatedAt   time.Time
}

// Validate checks usage counter invariants.
func (u UsageCounter) Validate() error {
	if u.ID == uuid.Nil {
		return fmt.Errorf("%w: usage id required", ErrInvalidArgument)
	}
	if u.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if u.PromotionID == uuid.Nil {
		return fmt.Errorf("%w: promotion_id required", ErrInvalidArgument)
	}
	return nil
}

// Simulation stores an evaluate dry-run for admin debugging.
type Simulation struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	RequestPayload map[string]any
	ResultPayload  map[string]any
	CreatedAt      time.Time
}
