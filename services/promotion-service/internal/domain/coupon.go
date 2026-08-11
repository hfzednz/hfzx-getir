package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CouponKind classifies coupon distribution.
type CouponKind string

const (
	CouponSingle   CouponKind = "single"
	CouponMulti    CouponKind = "multi"
	CouponPersonal CouponKind = "personal"
	CouponPublic   CouponKind = "public"
)

// Valid reports whether the coupon kind is recognized.
func (k CouponKind) Valid() bool {
	switch k {
	case CouponSingle, CouponMulti, CouponPersonal, CouponPublic:
		return true
	default:
		return false
	}
}

// Coupon binds a code to a promotion with usage limits.
type Coupon struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	PromotionID   uuid.UUID
	Code          string
	Kind          CouponKind
	MaxRedemptions int // 0 = unlimited (multi/public); single defaults to 1
	RedeemedCount int
	// PrincipalID for personal coupons (opaque user id).
	PrincipalID *uuid.UUID
	StartsAt    *time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks coupon invariants.
func (c Coupon) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: coupon id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if c.PromotionID == uuid.Nil {
		return fmt.Errorf("%w: promotion_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(c.Code) == "" {
		return fmt.Errorf("%w: coupon code required", ErrInvalidArgument)
	}
	if !c.Kind.Valid() {
		return fmt.Errorf("%w: invalid coupon kind %q", ErrInvalidArgument, c.Kind)
	}
	if c.Kind == CouponPersonal && (c.PrincipalID == nil || *c.PrincipalID == uuid.Nil) {
		return fmt.Errorf("%w: personal coupon requires principal_id", ErrInvalidArgument)
	}
	return nil
}

// IsValidAt reports whether the coupon can be used at now.
func (c Coupon) IsValidAt(now time.Time) bool {
	if c.StartsAt != nil && now.Before(*c.StartsAt) {
		return false
	}
	if c.EndsAt != nil && !now.Before(*c.EndsAt) {
		return false
	}
	max := c.MaxRedemptions
	if max == 0 && c.Kind == CouponSingle {
		max = 1
	}
	if max > 0 && c.RedeemedCount >= max {
		return false
	}
	return true
}

// CouponRedemption records a single redemption (idempotent by key).
type CouponRedemption struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CouponID       uuid.UUID
	PrincipalID    uuid.UUID
	IdempotencyKey string
	OrderRef       string // opaque external order/cart ref — not owned here
	DiscountMinor  int64
	Currency       string
	RedeemedAt     time.Time
	CreatedAt      time.Time
}
