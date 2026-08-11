package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// ValidateInput runs the ordered validation pipeline.
type ValidateInput struct {
	TenantID  uuid.UUID
	SessionID uuid.UUID
}

// Validate runs: customer → address/zone → inventory → price → coupon →
// restrictions → fraud → payment eligibility → duplicate → min order.
func (d *Deps) Validate(ctx context.Context, in ValidateInput) (domain.Session, error) {
	s, err := d.Sessions.GetByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.Session{}, err
	}
	switch s.Status {
	case domain.StatusStarted, domain.StatusValidating, domain.StatusReady, domain.StatusBlocked:
		// ok
	default:
		return domain.Session{}, fmt.Errorf("%w: cannot validate in status %s", domain.ErrInvalidTransition, s.Status)
	}

	if err := d.transition(&s, domain.StatusValidating); err != nil {
		return domain.Session{}, err
	}
	s.UpdatedAt = d.now()
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}

	cart, err := d.Cart.GetCart(ctx, s.TenantID, s.CartID)
	if err != nil {
		return domain.Session{}, err
	}

	issues := make([]domain.ValidationIssue, 0)
	var quote ports.QuoteResult
	var geo ports.GeofenceResult

	// 1. customer
	issues = append(issues, d.stepCustomer(ctx, s)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 2. address/zone
	addrIssues, g := d.stepAddressZone(ctx, s)
	geo = g
	issues = append(issues, addrIssues...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 3. inventory ATP
	issues = append(issues, d.stepInventory(ctx, s, cart)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 4. price refresh
	priceIssues, q := d.stepPrice(ctx, s, cart)
	quote = q
	issues = append(issues, priceIssues...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 5. coupon
	issues = append(issues, d.stepCoupon(ctx, s, quote)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 6. age/region restrictions
	issues = append(issues, d.stepRestrictions(ctx, s, cart)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 7. fraud/risk
	issues = append(issues, d.stepFraud(ctx, s, quote)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 8. payment eligibility
	issues = append(issues, d.stepPaymentEligibility(ctx, s, quote)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 9. duplicate detect
	issues = append(issues, d.stepDuplicate(ctx, s)...)
	if hasBlocking(issues) {
		return d.finishValidation(ctx, s, issues, quote)
	}

	// 10. min order
	issues = append(issues, d.stepMinOrder(s, quote, geo)...)

	return d.finishValidation(ctx, s, issues, quote)
}

func hasBlocking(issues []domain.ValidationIssue) bool {
	for _, i := range issues {
		if i.Blocking() {
			return true
		}
	}
	return false
}

func (d *Deps) finishValidation(ctx context.Context, s domain.Session, issues []domain.ValidationIssue, quote ports.QuoteResult) (domain.Session, error) {
	now := d.now()
	passed := !hasBlocking(issues)
	s.Validation = domain.ValidationResults{
		Passed:    passed,
		Issues:    issues,
		CheckedAt: now,
	}
	if quote.QuoteID != "" || quote.TotalMinor > 0 || quote.Currency != "" {
		s.Quote = domain.QuoteSnapshot{
			QuoteID:        quote.QuoteID,
			Currency:       quote.Currency,
			SubtotalMinor:  quote.SubtotalMinor,
			DiscountMinor:  quote.DiscountMinor,
			TaxMinor:       quote.TaxMinor,
			DeliveryMinor:  quote.DeliveryMinor,
			ServiceMinor:   quote.ServiceMinor,
			PackagingMinor: quote.PackagingMinor,
			TipMinor:       quote.TipMinor,
			TotalMinor:     quote.TotalMinor,
			QuotedAt:       quote.QuotedAt,
		}
		if s.Quote.QuotedAt.IsZero() {
			s.Quote.QuotedAt = now
		}
		if s.Currency == "" && quote.Currency != "" {
			s.Currency = quote.Currency
		}
	}

	next := domain.StatusBlocked
	if passed {
		next = domain.StatusReady
	}
	if err := d.transition(&s, next); err != nil {
		return domain.Session{}, err
	}
	if err := d.Sessions.Update(ctx, s); err != nil {
		return domain.Session{}, err
	}
	_ = d.appendEvent(ctx, s.ID, s.TenantID, domain.EventCheckoutValidated, map[string]any{
		"passed": passed,
		"status": string(s.Status),
		"issues": len(issues),
	})
	return s, nil
}

func (d *Deps) stepCustomer(ctx context.Context, s domain.Session) []domain.ValidationIssue {
	if d.Customer == nil {
		return nil
	}
	res, err := d.Customer.Check(ctx, ports.CustomerCheckRequest{
		TenantID: s.TenantID, PrincipalID: s.PrincipalID,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssueCustomerMissing, Message: err.Error(), Severity: "error",
		}}
	}
	if !res.Active {
		msg := res.Reason
		if msg == "" {
			msg = "customer inactive"
		}
		return []domain.ValidationIssue{{
			Code: domain.IssueCustomerInactive, Message: msg, Severity: "error",
		}}
	}
	return nil
}

func (d *Deps) stepAddressZone(ctx context.Context, s domain.Session) ([]domain.ValidationIssue, ports.GeofenceResult) {
	var geo ports.GeofenceResult
	if s.DeliveryOption == domain.DeliveryPickup {
		return nil, geo
	}
	if s.Address.Line1 == "" && s.Address.Lat == 0 && s.Address.Lng == 0 {
		return []domain.ValidationIssue{{
			Code: domain.IssueAddressMissing, Message: "delivery address required",
			Field: "address", Severity: "error",
		}}, geo
	}
	if s.DeliveryOption == domain.DeliveryScheduled {
		if s.Slot.SlotID == "" && s.Slot.StartsAt == nil {
			return []domain.ValidationIssue{{
				Code: domain.IssueSlotRequired, Message: "delivery slot required",
				Field: "slot", Severity: "error",
			}}, geo
		}
	}
	if d.Geofence == nil {
		return nil, geo
	}
	res, err := d.Geofence.CheckZone(ctx, ports.GeofenceRequest{
		TenantID: s.TenantID, CityID: s.CityID,
		Lat: s.Address.Lat, Lng: s.Address.Lng,
		DeliveryOption: s.DeliveryOption,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssueZoneUnavailable, Message: err.Error(), Severity: "error",
		}}, geo
	}
	geo = res
	if !res.InZone {
		msg := res.Reason
		if msg == "" {
			msg = "address outside delivery zone"
		}
		return []domain.ValidationIssue{{
			Code: domain.IssueZoneOutOfRange, Message: msg,
			Field: "address", Severity: "error",
		}}, geo
	}
	return nil, geo
}

func (d *Deps) stepInventory(ctx context.Context, s domain.Session, cart ports.CartView) []domain.ValidationIssue {
	if d.Inventory == nil {
		return nil
	}
	lines := make([]ports.ATPLineRequest, 0, len(cart.Lines))
	for _, l := range cart.Lines {
		lines = append(lines, ports.ATPLineRequest{
			VariantID: l.VariantID, SKUCode: l.SKUCode, Qty: l.Qty, WarehouseID: l.WarehouseID,
		})
	}
	res, err := d.Inventory.CheckATP(ctx, ports.ATPRequest{
		TenantID: s.TenantID, CityID: s.CityID, Lines: lines,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssueInventoryUnavailable, Message: err.Error(), Severity: "error",
		}}
	}
	if res.AllAvailable {
		return nil
	}
	issues := make([]domain.ValidationIssue, 0)
	for _, lr := range res.Lines {
		if lr.Available {
			continue
		}
		msg := lr.Reason
		if msg == "" {
			msg = "insufficient stock"
		}
		issues = append(issues, domain.ValidationIssue{
			Code: domain.IssueInventoryInsufficient, Message: msg, Severity: "error",
			Meta: map[string]any{
				"variantId":    lr.VariantID.String(),
				"availableQty": lr.AvailableQty,
			},
		})
	}
	if len(issues) == 0 {
		issues = append(issues, domain.ValidationIssue{
			Code: domain.IssueInventoryInsufficient, Message: "inventory not available", Severity: "error",
		})
	}
	return issues
}

func (d *Deps) stepPrice(ctx context.Context, s domain.Session, cart ports.CartView) ([]domain.ValidationIssue, ports.QuoteResult) {
	var empty ports.QuoteResult
	if d.Pricing == nil {
		return []domain.ValidationIssue{{
			Code: domain.IssuePriceFailed, Message: "pricing client not configured", Severity: "error",
		}}, empty
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
		return []domain.ValidationIssue{{
			Code: domain.IssuePriceFailed, Message: err.Error(), Severity: "error",
		}}, empty
	}
	if q.TotalMinor < 0 {
		return []domain.ValidationIssue{{
			Code: domain.IssuePriceFailed, Message: "negative total", Severity: "error",
		}}, empty
	}
	return nil, q
}

func (d *Deps) stepCoupon(ctx context.Context, s domain.Session, quote ports.QuoteResult) []domain.ValidationIssue {
	if len(s.CouponCodes) == 0 || d.Promo == nil {
		return nil
	}
	res, err := d.Promo.Validate(ctx, ports.PromoRequest{
		TenantID: s.TenantID, PrincipalID: s.PrincipalID,
		Codes: s.CouponCodes, SubtotalMinor: quote.SubtotalMinor,
		Currency: s.Currency, CityID: s.CityID,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssueCouponInvalid, Message: err.Error(), Severity: "error",
		}}
	}
	if !res.Valid {
		msg := res.Reason
		if msg == "" {
			msg = "coupon not eligible"
		}
		return []domain.ValidationIssue{{
			Code: domain.IssueCouponIneligible, Message: msg, Severity: "error",
		}}
	}
	return nil
}

func (d *Deps) stepRestrictions(ctx context.Context, s domain.Session, cart ports.CartView) []domain.ValidationIssue {
	for _, l := range cart.Lines {
		if !l.AgeRestricted {
			continue
		}
		if d.Customer == nil {
			return []domain.ValidationIssue{{
				Code: domain.IssueAgeRestricted, Message: "age-restricted item requires verified customer",
				Severity: "error",
				Meta:     map[string]any{"variantId": l.VariantID.String()},
			}}
		}
		res, err := d.Customer.Check(ctx, ports.CustomerCheckRequest{
			TenantID: s.TenantID, PrincipalID: s.PrincipalID,
		})
		if err != nil || res.Age < 18 {
			return []domain.ValidationIssue{{
				Code: domain.IssueAgeRestricted, Message: "customer does not meet age restriction",
				Severity: "error",
			}}
		}
	}
	return nil
}

func (d *Deps) stepFraud(ctx context.Context, s domain.Session, quote ports.QuoteResult) []domain.ValidationIssue {
	if d.Fraud == nil {
		return nil
	}
	res, err := d.Fraud.Score(ctx, ports.FraudRequest{
		TenantID: s.TenantID, PrincipalID: s.PrincipalID, CheckoutID: s.ID,
		TotalMinor: quote.TotalMinor, Currency: s.Currency, CityID: s.CityID,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssueFraudHighRisk, Message: err.Error(), Severity: "error",
		}}
	}
	if res.Decision == "block" {
		msg := res.Reason
		if msg == "" {
			msg = "fraud risk too high"
		}
		return []domain.ValidationIssue{{
			Code: domain.IssueFraudHighRisk, Message: msg, Severity: "error",
			Meta: map[string]any{"score": res.Score},
		}}
	}
	return nil
}

func (d *Deps) stepPaymentEligibility(ctx context.Context, s domain.Session, quote ports.QuoteResult) []domain.ValidationIssue {
	if d.PayElig == nil {
		return nil
	}
	res, err := d.PayElig.Check(ctx, ports.PaymentEligibilityRequest{
		TenantID: s.TenantID, PrincipalID: s.PrincipalID,
		TotalMinor: quote.TotalMinor, Currency: s.Currency,
	})
	if err != nil {
		return []domain.ValidationIssue{{
			Code: domain.IssuePaymentIneligible, Message: err.Error(), Severity: "error",
		}}
	}
	if !res.Eligible {
		msg := res.Reason
		if msg == "" {
			msg = "no eligible payment methods"
		}
		return []domain.ValidationIssue{{
			Code: domain.IssuePaymentIneligible, Message: msg, Severity: "error",
		}}
	}
	return nil
}

func (d *Deps) stepDuplicate(ctx context.Context, s domain.Session) []domain.ValidationIssue {
	// Soft duplicate: another completed session for same cart recently.
	sessions, _, err := d.Sessions.List(ctx, ports.SessionFilter{
		TenantID: s.TenantID, Limit: 50,
	})
	if err != nil {
		return nil
	}
	for _, other := range sessions {
		if other.ID == s.ID {
			continue
		}
		if other.CartID == s.CartID && other.Status == domain.StatusCompleted && other.OrderID != "" {
			return []domain.ValidationIssue{{
				Code: domain.IssueDuplicateOrder, Message: "cart already completed as an order",
				Severity: "error",
				Meta:     map[string]any{"orderId": other.OrderID},
			}}
		}
	}
	return nil
}

func (d *Deps) stepMinOrder(s domain.Session, quote ports.QuoteResult, geo ports.GeofenceResult) []domain.ValidationIssue {
	min := geo.MinOrderMinor
	if min <= 0 {
		min = d.MinOrderMinor
	}
	if min <= 0 {
		return nil
	}
	if quote.SubtotalMinor < min {
		return []domain.ValidationIssue{{
			Code: domain.IssueMinOrderNotMet,
			Message: fmt.Sprintf("subtotal %d below minimum %d", quote.SubtotalMinor, min),
			Severity: "error",
			Meta: map[string]any{"minOrderMinor": min, "subtotalMinor": quote.SubtotalMinor},
		}}
	}
	return nil
}
