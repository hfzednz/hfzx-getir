package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// QuoteLine is a priced cart line after waterfall + dynamic + promo allocation.
type QuoteLine struct {
	VariantID         uuid.UUID
	Qty               int
	UnitPriceMinor    int64 // after dynamic
	BaseUnitMinor     int64 // waterfall unit before dynamic
	DynamicAdjMinor   int64 // unit delta (can be negative conceptually via lower unit)
	LineSubtotalMinor int64 // unit * qty before line discount
	DiscountMinor     int64
	LineTotalMinor    int64 // line subtotal - discount
	ResolvedScope     PriceScope
	PriceEntryID      uuid.UUID
}

// PromoDiscount is a discount attributed to a promotion evaluation.
type PromoDiscount struct {
	PromotionID   string
	Code          string
	DiscountMinor int64
	Description   string
}

// Quote is the assembled cart/checkout price response (all minor units).
type Quote struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CartID         *uuid.UUID
	Currency       string
	Lines          []QuoteLine
	SubtotalMinor  int64
	DiscountMinor  int64
	TaxMinor       int64
	DeliveryMinor  int64
	ServiceMinor   int64
	PackagingMinor int64
	TipMinor       int64
	TotalMinor     int64
	Promos         []PromoDiscount
	TaxRuleCode    string
	Simulated      bool
	QuotedAt       time.Time
}

// Validate checks quote money invariants.
func (q Quote) Validate() error {
	if q.ID == uuid.Nil {
		return fmt.Errorf("%w: quote id required", ErrInvalidArgument)
	}
	if q.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if _, err := NewMoney(0, q.Currency); err != nil {
		return err
	}
	for _, v := range []int64{
		q.SubtotalMinor, q.DiscountMinor, q.TaxMinor, q.DeliveryMinor,
		q.ServiceMinor, q.PackagingMinor, q.TipMinor, q.TotalMinor,
	} {
		if v < 0 {
			return fmt.Errorf("%w: negative quote component", ErrNegativeMoney)
		}
	}
	if len(q.Lines) == 0 {
		return fmt.Errorf("%w", ErrEmptyQuote)
	}
	if q.QuotedAt.IsZero() {
		return fmt.Errorf("%w: quoted_at required", ErrInvalidArgument)
	}
	return nil
}

// RecomputeTotals derives subtotal/discount/total from lines + fees + tax.
func (q *Quote) RecomputeTotals() {
	var sub, disc int64
	for i := range q.Lines {
		ln := &q.Lines[i]
		ln.LineSubtotalMinor = ln.UnitPriceMinor * int64(ln.Qty)
		if ln.DiscountMinor < 0 {
			ln.DiscountMinor = 0
		}
		ln.LineTotalMinor = ln.LineSubtotalMinor - ln.DiscountMinor
		if ln.LineTotalMinor < 0 {
			ln.LineTotalMinor = 0
		}
		sub += ln.LineSubtotalMinor
		disc += ln.DiscountMinor
	}
	q.SubtotalMinor = sub
	q.DiscountMinor = disc
	taxable := sub - disc
	if taxable < 0 {
		taxable = 0
	}
	// TaxMinor set by caller via TaxCalculate; preserve if already set for inclusive flows
	q.TotalMinor = taxable + q.TaxMinor + q.DeliveryMinor + q.ServiceMinor + q.PackagingMinor + q.TipMinor
}

// QuoteAudit is a persisted quote snapshot for debugging / simulate.
type QuoteAudit struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	QuoteID   uuid.UUID
	CartID    *uuid.UUID
	Simulated bool
	Request   map[string]any
	Response  map[string]any
	CreatedAt time.Time
}
