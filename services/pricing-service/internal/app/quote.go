package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// QuoteLineInput is a cart line for quote assembly.
type QuoteLineInput struct {
	VariantID uuid.UUID
	Qty       int
}

// QuoteCartInput assembles a full cart quote.
type QuoteCartInput struct {
	TenantID       uuid.UUID
	CartID         *uuid.UUID
	Currency       string
	RegionID       *uuid.UUID
	WarehouseID    *uuid.UUID
	CustomerID     *uuid.UUID
	VIPID          *uuid.UUID
	CorporateID    *uuid.UUID
	CouponCodes    []string
	Lines          []QuoteLineInput
	DeliveryMinor  int64
	ServiceMinor   int64
	PackagingMinor int64
	TipMinor       int64
	Simulate       bool
}

// QuoteCart resolves prices, applies dynamic, promo evaluate, tax, and fees.
func (d *Deps) QuoteCart(ctx context.Context, in QuoteCartInput) (domain.Quote, error) {
	if in.TenantID == uuid.Nil {
		return domain.Quote{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if _, err := domain.NewMoney(0, currency); err != nil {
		return domain.Quote{}, err
	}
	if len(in.Lines) == 0 {
		return domain.Quote{}, fmt.Errorf("%w", domain.ErrEmptyQuote)
	}
	for _, fee := range []int64{in.DeliveryMinor, in.ServiceMinor, in.PackagingMinor, in.TipMinor} {
		if fee < 0 {
			return domain.Quote{}, fmt.Errorf("%w: fee", domain.ErrNegativeMoney)
		}
	}

	now := d.now()
	quote := domain.Quote{
		ID: d.newID(), TenantID: in.TenantID, CartID: in.CartID,
		Currency: currency, DeliveryMinor: in.DeliveryMinor,
		ServiceMinor: in.ServiceMinor, PackagingMinor: in.PackagingMinor,
		TipMinor: in.TipMinor, Simulated: in.Simulate, QuotedAt: now,
		Lines: make([]domain.QuoteLine, 0, len(in.Lines)),
	}

	promoLines := make([]ports.PromoLineInput, 0, len(in.Lines))
	for _, ln := range in.Lines {
		if ln.VariantID == uuid.Nil || ln.Qty <= 0 {
			return domain.Quote{}, fmt.Errorf("%w: line variant_id and qty>0 required", domain.ErrInvalidArgument)
		}
		resolved, err := d.GetPrice(ctx, GetPriceInput{
			TenantID: in.TenantID, VariantID: ln.VariantID, Currency: currency,
			RegionID: in.RegionID, WarehouseID: in.WarehouseID, CustomerID: in.CustomerID,
			VIPID: in.VIPID, CorporateID: in.CorporateID, At: &now,
		})
		if err != nil {
			return domain.Quote{}, err
		}
		dyn, err := d.ApplyDynamic(ctx, ApplyDynamicInput{
			TenantID: in.TenantID, VariantID: ln.VariantID,
			UnitMinor: resolved.AmountMinor, Currency: currency, WarehouseID: in.WarehouseID,
		})
		if err != nil {
			return domain.Quote{}, err
		}
		ql := domain.QuoteLine{
			VariantID: ln.VariantID, Qty: ln.Qty,
			UnitPriceMinor: dyn.UnitMinor, BaseUnitMinor: dyn.BaseUnitMinor,
			DynamicAdjMinor: dyn.AdjustmentMinor,
			ResolvedScope:   resolved.Scope, PriceEntryID: resolved.Entry.ID,
		}
		quote.Lines = append(quote.Lines, ql)
		promoLines = append(promoLines, ports.PromoLineInput{
			VariantID: ln.VariantID, Qty: ln.Qty,
			UnitPriceMinor: dyn.UnitMinor, LineTotalMinor: dyn.UnitMinor * int64(ln.Qty),
		})
	}

	quote.RecomputeTotals()

	// Promo evaluate (no local promo storage)
	if d.Promo != nil {
		eval, err := d.Promo.Evaluate(ctx, ports.PromoEvaluateRequest{
			TenantID: in.TenantID, CustomerID: in.CustomerID, Currency: currency,
			CouponCodes: in.CouponCodes, Lines: promoLines, SubtotalMinor: quote.SubtotalMinor,
		})
		if err != nil {
			return domain.Quote{}, err
		}
		allocatePromoDiscounts(&quote, eval)
	}

	quote.RecomputeTotals()

	taxable := quote.SubtotalMinor - quote.DiscountMinor
	if taxable < 0 {
		taxable = 0
	}
	taxRes, err := d.TaxCalculate(ctx, TaxCalculateInput{
		TenantID: in.TenantID, RegionID: in.RegionID, BaseMinor: taxable, Currency: currency,
	})
	if err != nil {
		return domain.Quote{}, err
	}
	quote.TaxMinor = taxRes.TaxMinor
	quote.TaxRuleCode = taxRes.RuleCode
	quote.RecomputeTotals()

	if err := quote.Validate(); err != nil {
		return domain.Quote{}, err
	}

	if d.Audits != nil {
		_ = d.Audits.Create(ctx, domain.QuoteAudit{
			ID: d.newID(), TenantID: in.TenantID, QuoteID: quote.ID,
			CartID: in.CartID, Simulated: in.Simulate,
			Request: map[string]any{
				"currency": currency, "lineCount": len(in.Lines),
				"coupons": in.CouponCodes, "simulate": in.Simulate,
			},
			Response: map[string]any{
				"subtotalMinor": quote.SubtotalMinor, "discountMinor": quote.DiscountMinor,
				"taxMinor": quote.TaxMinor, "totalMinor": quote.TotalMinor,
			},
			CreatedAt: now,
		})
	}

	if !in.Simulate {
		d.emit(ctx, in.TenantID, quote.ID, domain.EventQuoteCreated, map[string]any{
			"currency": currency, "totalMinor": quote.TotalMinor, "subtotalMinor": quote.SubtotalMinor,
		})
	}
	return quote, nil
}

// SimulateQuote is QuoteCart with Simulated=true (no QuoteCreated event).
func (d *Deps) SimulateQuote(ctx context.Context, in QuoteCartInput) (domain.Quote, error) {
	in.Simulate = true
	return d.QuoteCart(ctx, in)
}

func allocatePromoDiscounts(quote *domain.Quote, eval ports.PromoEvaluateResult) {
	quote.Promos = make([]domain.PromoDiscount, 0, len(eval.Discounts))
	remaining := eval.DiscountMinor
	if remaining < 0 {
		remaining = 0
	}

	// Prefer line-allocated discounts
	allocated := int64(0)
	for _, d := range eval.Discounts {
		quote.Promos = append(quote.Promos, domain.PromoDiscount{
			PromotionID: d.PromotionID, Code: d.Code,
			DiscountMinor: d.DiscountMinor, Description: d.Description,
		})
		if d.VariantID != nil && d.DiscountMinor > 0 {
			for i := range quote.Lines {
				if quote.Lines[i].VariantID == *d.VariantID {
					quote.Lines[i].DiscountMinor += d.DiscountMinor
					allocated += d.DiscountMinor
					break
				}
			}
		}
	}

	// Distribute remaining (cart-level) proportional to line subtotals
	cartLevel := remaining - allocated
	if cartLevel <= 0 && remaining > 0 && allocated == 0 {
		cartLevel = remaining
	}
	if cartLevel > 0 && quote.SubtotalMinor > 0 {
		var distributed int64
		for i := range quote.Lines {
			share := quote.Lines[i].UnitPriceMinor * int64(quote.Lines[i].Qty) * cartLevel / quote.SubtotalMinor
			quote.Lines[i].DiscountMinor += share
			distributed += share
		}
		// fix rounding on last line
		diff := cartLevel - distributed
		if diff != 0 && len(quote.Lines) > 0 {
			last := &quote.Lines[len(quote.Lines)-1]
			last.DiscountMinor += diff
			if last.DiscountMinor < 0 {
				last.DiscountMinor = 0
			}
		}
	}
}
