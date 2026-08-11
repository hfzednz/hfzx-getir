package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// RefreshQuoteInput refreshes the cart quote via PricingClient.
type RefreshQuoteInput struct {
	TenantID    uuid.UUID
	CartID      uuid.UUID
	SoftReserve bool // optional: soft-reserve after quote
}

// RefreshQuote calls PricingClient.Quote and stores the snapshot.
// When SoftReserve is true, also soft-reserves inventory after quote.
func (d *Deps) RefreshQuote(ctx context.Context, in RefreshQuoteInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	if len(c.Lines) == 0 {
		return domain.Cart{}, fmt.Errorf("%w", domain.ErrEmptyCart)
	}
	if d.Pricing == nil {
		return domain.Cart{}, fmt.Errorf("%w: pricing client required", domain.ErrInvariant)
	}

	lines := make([]ports.QuoteLineInput, 0, len(c.Lines))
	for _, l := range c.Lines {
		lines = append(lines, ports.QuoteLineInput{VariantID: l.VariantID, Qty: l.Qty})
	}
	qr, err := d.Pricing.Quote(ctx, ports.QuoteRequest{
		TenantID:    c.TenantID,
		CartID:      c.ID,
		Currency:    c.Currency,
		CityID:      c.CityID,
		CouponCodes: c.CouponCodes(),
		Lines:       lines,
	})
	if err != nil {
		return domain.Cart{}, err
	}
	now := d.now()
	quotedAt := qr.QuotedAt
	if quotedAt.IsZero() {
		quotedAt = now
	}
	snap := &domain.QuoteSnapshot{
		QuoteID:        qr.QuoteID,
		Currency:       qr.Currency,
		SubtotalMinor:  qr.SubtotalMinor,
		DiscountMinor:  qr.DiscountMinor,
		TaxMinor:       qr.TaxMinor,
		DeliveryMinor:  qr.DeliveryMinor,
		ServiceMinor:   qr.ServiceMinor,
		PackagingMinor: qr.PackagingMinor,
		TipMinor:       qr.TipMinor,
		TotalMinor:     qr.TotalMinor,
		LineQuotes:     qr.LineQuotes,
		QuotedAt:       quotedAt,
	}
	if snap.Currency == "" {
		snap.Currency = c.Currency
	}
	if err := snap.Validate(); err != nil {
		return domain.Cart{}, err
	}
	// Sync coupon discount preview from quote when present.
	if snap.DiscountMinor > 0 && len(c.Coupons) > 0 {
		c.Coupons[0].DiscountMinor = snap.DiscountMinor
	}
	c.Quote = snap
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventQuoteRefreshed, map[string]any{
		"quoteId":    snap.QuoteID.String(),
		"totalMinor": snap.TotalMinor,
	})

	if in.SoftReserve {
		c, err = d.SoftReserveLines(ctx, SoftReserveLinesInput{
			TenantID:       c.TenantID,
			CartID:         c.ID,
			IdempotencyKey: "quote-" + snap.QuoteID.String(),
		})
		if err != nil {
			return domain.Cart{}, err
		}
	}
	return c, nil
}
