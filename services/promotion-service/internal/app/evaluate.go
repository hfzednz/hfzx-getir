package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// EvaluateCart evaluates promotions against a cart context (core engine).
func (d *Deps) EvaluateCart(ctx context.Context, eval domain.EvaluateContext) (domain.EvaluateResult, error) {
	if eval.TenantID == uuid.Nil {
		return domain.EvaluateResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(eval.Currency))
	if len(currency) != 3 {
		return domain.EvaluateResult{}, fmt.Errorf("%w: currency required", domain.ErrInvalidArgument)
	}
	now := eval.Now
	if now.IsZero() {
		now = d.now()
	}

	campaigns, err := d.Campaigns.ListActive(ctx, eval.TenantID, now)
	if err != nil {
		return domain.EvaluateResult{}, err
	}

	// Coupon-bound promotions (may be outside auto-apply set).
	couponPromoIDs := map[uuid.UUID]string{}
	for _, code := range eval.CouponCodes {
		c, err := d.Coupons.GetByCode(ctx, eval.TenantID, code)
		if err != nil {
			continue
		}
		if !c.IsValidAt(now) {
			continue
		}
		if c.Kind == domain.CouponPersonal && c.PrincipalID != nil && eval.PrincipalID != *c.PrincipalID {
			continue
		}
		couponPromoIDs[c.PromotionID] = strings.ToUpper(strings.TrimSpace(code))
	}

	var candidates []domain.Candidate
	seen := map[uuid.UUID]struct{}{}

	for _, camp := range campaigns {
		promos, err := d.Promotions.ListByCampaign(ctx, eval.TenantID, camp.ID)
		if err != nil {
			return domain.EvaluateResult{}, err
		}
		for _, p := range promos {
			cand, ok, err := d.buildCandidate(ctx, eval, camp, p, couponPromoIDs, currency)
			if err != nil {
				return domain.EvaluateResult{}, err
			}
			if !ok {
				continue
			}
			seen[p.ID] = struct{}{}
			candidates = append(candidates, cand)
		}
	}

	// Coupon promotions whose campaign might still be active but listed separately —
	// if promotion not yet seen, try load it (coupon-only path when campaign active).
	for pid, code := range couponPromoIDs {
		if _, ok := seen[pid]; ok {
			continue
		}
		p, err := d.Promotions.GetByID(ctx, eval.TenantID, pid)
		if err != nil {
			continue
		}
		camp, err := d.Campaigns.GetByID(ctx, eval.TenantID, p.CampaignID)
		if err != nil || !camp.IsActiveAt(now) {
			continue
		}
		codes := map[uuid.UUID]string{pid: code}
		cand, ok, err := d.buildCandidate(ctx, eval, camp, p, codes, currency)
		if err != nil {
			return domain.EvaluateResult{}, err
		}
		if ok {
			candidates = append(candidates, cand)
		}
	}

	winners := domain.ResolveConflicts(candidates)
	var total, shipDisc int64
	for _, w := range winners {
		total += w.AmountMinor
		if w.Type == domain.PromoFreeShip {
			shipDisc += eval.ShippingMinor
		}
	}
	return domain.EvaluateResult{
		Discounts:             winners,
		TotalDiscountMinor:    total,
		Currency:              currency,
		ShippingDiscountMinor: shipDisc,
	}, nil
}

func (d *Deps) buildCandidate(
	ctx context.Context,
	eval domain.EvaluateContext,
	camp domain.Campaign,
	p domain.Promotion,
	couponPromoIDs map[uuid.UUID]string,
	currency string,
) (domain.Candidate, bool, error) {
	rule, err := d.Rules.GetByPromotionID(ctx, eval.TenantID, p.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			rule = domain.Rule{Priority: p.Priority, PromotionID: p.ID, TenantID: eval.TenantID}
		} else {
			return domain.Candidate{}, false, err
		}
	}
	if !rule.MatchesSegments(eval.SegmentIDs) {
		return domain.Candidate{}, false, nil
	}

	// Usage limits
	if rule.PerUserLimit > 0 && eval.PrincipalID != uuid.Nil {
		if u, err := d.Usage.Get(ctx, eval.TenantID, p.ID, domain.UsageUser, eval.PrincipalID.String()); err == nil {
			if u.Count >= rule.PerUserLimit {
				return domain.Candidate{}, false, nil
			}
		}
	}
	if rule.GlobalLimit > 0 {
		if u, err := d.Usage.Get(ctx, eval.TenantID, p.ID, domain.UsageGlobal, "global"); err == nil {
			if u.Count >= rule.GlobalLimit {
				return domain.Candidate{}, false, nil
			}
		}
	}
	if rule.PerDeviceLimit > 0 && eval.DeviceID != "" {
		if u, err := d.Usage.Get(ctx, eval.TenantID, p.ID, domain.UsageDevice, eval.DeviceID); err == nil {
			if u.Count >= rule.PerDeviceLimit {
				return domain.Candidate{}, false, nil
			}
		}
	}

	var matched []domain.CartLine
	for _, line := range eval.Lines {
		if rule.MatchesLine(line.VariantID, line.CategoryID, line.BrandID) {
			matched = append(matched, line)
		}
	}
	if len(matched) == 0 && p.Type != domain.PromoFreeShip {
		return domain.Candidate{}, false, nil
	}
	if rule.MinQty > 0 {
		var qty int
		for _, l := range matched {
			qty += l.Quantity
		}
		if qty < rule.MinQty {
			return domain.Candidate{}, false, nil
		}
	}

	amount, shipPart, appliedIDs := domain.ComputeDiscount(p, matched, eval.ShippingMinor, currency)
	if amount == 0 && shipPart == 0 && p.Type != domain.PromoGift {
		return domain.Candidate{}, false, nil
	}

	priority := rule.Priority
	if priority == 0 {
		priority = p.Priority
	}
	couponCode := couponPromoIDs[p.ID]
	desc := p.Name
	dl := domain.DiscountLine{
		PromotionID:    p.ID,
		CampaignID:     camp.ID,
		Type:           p.Type,
		AmountMinor:    amount,
		Currency:       currency,
		Description:    desc,
		StackGroup:     rule.StackGroup,
		Priority:       priority,
		AppliedLineIDs: appliedIDs,
		CouponCode:     couponCode,
	}
	return domain.Candidate{
		Promotion: p,
		Rule:      rule,
		Campaign:  camp,
		Discount:  dl,
	}, true, nil
}

// CommitEvaluation increments usage counters for applied promotions (optional post-checkout).
func (d *Deps) CommitEvaluation(ctx context.Context, tenantID, principalID uuid.UUID, deviceID, orderRef string, promotionIDs []uuid.UUID) error {
	for _, pid := range promotionIDs {
		if _, err := d.Usage.Increment(ctx, tenantID, pid, domain.UsageGlobal, "global"); err != nil {
			return err
		}
		if principalID != uuid.Nil {
			if _, err := d.Usage.Increment(ctx, tenantID, pid, domain.UsageUser, principalID.String()); err != nil {
				return err
			}
		}
		if deviceID != "" {
			if _, err := d.Usage.Increment(ctx, tenantID, pid, domain.UsageDevice, deviceID); err != nil {
				return err
			}
		}
		if orderRef != "" {
			if _, err := d.Usage.Increment(ctx, tenantID, pid, domain.UsageOrder, orderRef); err != nil {
				return err
			}
		}
		d.emit(ctx, tenantID, pid, domain.EventPromotionApplied, map[string]any{
			"orderRef": orderRef, "principalId": principalID.String(),
		})
	}
	return nil
}
