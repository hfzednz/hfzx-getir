package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// RefreshQuoteInput refreshes the pricing preview without full validation.
type RefreshQuoteInput struct {
	TenantID  uuid.UUID
	SessionID uuid.UUID
}

// RefreshQuote updates the quote snapshot via PricingClient.
func (d *Deps) RefreshQuote(ctx context.Context, in RefreshQuoteInput) (domain.Session, error) {
	s, err := d.Sessions.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.Session{}, err
	}
	if s.Status.IsTerminal() {
		return domain.Session{}, fmt.Errorf("%w: cannot refresh quote in status %s", domain.ErrInvalidTransition, s.Status)
	}
	cart, err := d.Cart.GetCart(ctx, s.TenantID, s.CartID)
	if err != nil {
		return domain.Session{}, err
	}
	if d.Pricing == nil {
		return domain.Session{}, fmt.Errorf("%w: pricing client not configured", domain.ErrInvariant)
	}
	qlines := make([]ports.QuoteLineRequest, 0, len(cart.Lines))
	for _, l := range cart.Lines {
		qlines = append(qlines, ports.QuoteLineRequest{
			VariantID: l.VariantID, SKUCode: l.SKUCode, Qty: l.Qty,
		})
	}
	q, err := d.Pricing.Quote(ctx, ports.QuoteRequest{
		TenantID: s.TenantID, CartID: s.CartID, CheckoutID: s.ID,
		Currency: s.Currency, CityID: s.CityID, DeliveryOption: s.DeliveryOption,
		CouponCodes: s.CouponCodes, TipMinor: s.TipMinor, Lines: qlines,
	})
	if err != nil {
		return domain.Session{}, err
	}
	now := d.now()
	s.Quote = domain.QuoteSnapshot{
		QuoteID:        q.QuoteID,
		Currency:       q.Currency,
		SubtotalMinor:  q.SubtotalMinor,
		DiscountMinor:  q.DiscountMinor,
		TaxMinor:       q.TaxMinor,
		DeliveryMinor:  q.DeliveryMinor,
		ServiceMinor:   q.ServiceMinor,
		PackagingMinor: q.PackagingMinor,
		TipMinor:       q.TipMinor,
		TotalMinor:     q.TotalMinor,
		QuotedAt:       q.QuotedAt,
		LineCount:      len(cart.Lines),
	}
	if s.Quote.QuotedAt.IsZero() {
		s.Quote.QuotedAt = now
	}
	if s.Status == domain.StatusReady {
		// Quote refresh invalidates ready — must re-validate.
		s.Status = domain.StatusStarted
		s.Validation = domain.ValidationResults{}
		s.Version++
	}
	s.UpdatedAt = now
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}
	return s, nil
}
