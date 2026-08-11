package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// QuoteSnapshot stores the last PricingClient quote (minor units).
type QuoteSnapshot struct {
	QuoteID         uuid.UUID
	Currency        string
	SubtotalMinor   int64
	DiscountMinor   int64
	TaxMinor        int64
	DeliveryMinor   int64
	ServiceMinor    int64
	PackagingMinor  int64
	TipMinor        int64
	TotalMinor      int64
	LineQuotes      []LineQuote
	QuotedAt        time.Time
}

// LineQuote is per-line pricing from PricingClient.
type LineQuote struct {
	VariantID       uuid.UUID `json:"variantId"`
	Qty             int       `json:"qty"`
	UnitPriceMinor  int64     `json:"unitPriceMinor"`
	LineTotalMinor  int64     `json:"lineTotalMinor"`
	DiscountMinor   int64     `json:"discountMinor"`
}

// Validate checks quote money invariants.
func (q QuoteSnapshot) Validate() error {
	if q.QuoteID == uuid.Nil {
		return fmt.Errorf("%w: quote_id required", ErrInvalidArgument)
	}
	if len(q.Currency) != 3 {
		return fmt.Errorf("%w: currency required", ErrInvalidArgument)
	}
	for _, v := range []int64{
		q.SubtotalMinor, q.DiscountMinor, q.TaxMinor, q.DeliveryMinor,
		q.ServiceMinor, q.PackagingMinor, q.TipMinor, q.TotalMinor,
	} {
		if v < 0 {
			return fmt.Errorf("%w: negative quote component", ErrNegativeMoney)
		}
	}
	if q.QuotedAt.IsZero() {
		return fmt.Errorf("%w: quoted_at required", ErrInvalidArgument)
	}
	return nil
}

// AppliedCoupon is a preview coupon code on the cart.
type AppliedCoupon struct {
	Code          string
	DiscountMinor int64
	AppliedAt     time.Time
	Metadata      map[string]any
}

// Validate checks coupon invariants.
func (c AppliedCoupon) Validate() error {
	if c.Code == "" {
		return fmt.Errorf("%w: coupon code required", ErrInvalidArgument)
	}
	if c.DiscountMinor < 0 {
		return fmt.Errorf("%w: discount", ErrNegativeMoney)
	}
	return nil
}

// ReservationRef is an opaque InventoryClient soft-reserve reference.
type ReservationRef struct {
	ID             uuid.UUID
	ReservationRef string
	IdempotencyKey string
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	ReleasedAt     *time.Time
}

// Active reports whether the reservation is still held.
func (r ReservationRef) Active() bool {
	return r.ReleasedAt == nil && r.ReservationRef != ""
}

// SavedCart is a named save-for-later snapshot.
type SavedCart struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	PrincipalID  uuid.UUID
	SourceCartID *uuid.UUID
	Name         string
	Snapshot     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// WishlistLink links an external wishlist item into a cart.
type WishlistLink struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CartID         uuid.UUID
	WishlistID     uuid.UUID
	WishlistItemID *uuid.UUID
	VariantID      uuid.UUID
	CreatedAt      time.Time
}
