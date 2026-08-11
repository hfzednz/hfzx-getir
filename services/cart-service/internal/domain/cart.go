package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CartStatus is the cart aggregate lifecycle.
type CartStatus string

const (
	CartStatusActive    CartStatus = "active"
	CartStatusAbandoned CartStatus = "abandoned"
	CartStatusConverted CartStatus = "converted"
	CartStatusMerged    CartStatus = "merged"
)

// Valid reports whether the status is recognized.
func (s CartStatus) Valid() bool {
	switch s {
	case CartStatusActive, CartStatusAbandoned, CartStatusConverted, CartStatusMerged:
		return true
	default:
		return false
	}
}

// IsMutable reports whether lines/coupons may change.
func (s CartStatus) IsMutable() bool {
	return s == CartStatusActive || s == CartStatusAbandoned
}

// Cart is the shopping cart aggregate (guest or principal).
type Cart struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	GuestToken      string
	PrincipalID     *uuid.UUID
	CityID          *uuid.UUID
	Status          CartStatus
	Currency        string
	Lines           []CartLine
	Coupons         []AppliedCoupon
	Quote           *QuoteSnapshot
	ReservationRefs []ReservationRef
	Version         int64
	MergedIntoID    *uuid.UUID
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	AbandonedAt     *time.Time
	ConvertedAt     *time.Time
	MergedAt        *time.Time
}

// Validate checks structural cart invariants.
func (c Cart) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: cart id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !c.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, c.Status)
	}
	if len(c.Currency) != 3 {
		return fmt.Errorf("%w: currency required", ErrInvalidArgument)
	}
	hasGuest := c.GuestToken != ""
	hasPrincipal := c.PrincipalID != nil && *c.PrincipalID != uuid.Nil
	if !hasGuest && !hasPrincipal && c.Status == CartStatusActive {
		return fmt.Errorf("%w", ErrOwnerRequired)
	}
	for _, l := range c.Lines {
		if err := l.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// RequireActive returns ErrCartInactive when the cart cannot be mutated.
func (c Cart) RequireActive() error {
	if !c.Status.IsMutable() {
		return fmt.Errorf("%w: status %s", ErrCartInactive, c.Status)
	}
	return nil
}

// LineByVariant finds a line by opaque variant id.
func (c Cart) LineByVariant(variantID uuid.UUID) (CartLine, bool) {
	for _, l := range c.Lines {
		if l.VariantID == variantID {
			return l, true
		}
	}
	return CartLine{}, false
}

// LineByID finds a line by id.
func (c Cart) LineByID(lineID uuid.UUID) (CartLine, bool) {
	for _, l := range c.Lines {
		if l.ID == lineID {
			return l, true
		}
	}
	return CartLine{}, false
}

// HasCoupon reports whether code is applied.
func (c Cart) HasCoupon(code string) bool {
	for _, cp := range c.Coupons {
		if cp.Code == code {
			return true
		}
	}
	return false
}

// CouponCodes returns applied coupon codes.
func (c Cart) CouponCodes() []string {
	out := make([]string, 0, len(c.Coupons))
	for _, cp := range c.Coupons {
		out = append(out, cp.Code)
	}
	return out
}
