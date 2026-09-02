package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// GenerateCouponInput creates a coupon code bound to a promotion.
type GenerateCouponInput struct {
	TenantID       uuid.UUID
	PromotionID    uuid.UUID
	Code           string
	Kind           domain.CouponKind
	MaxRedemptions int
	PrincipalID    *uuid.UUID
	StartsAt       *time.Time
	EndsAt         *time.Time
}

// GenerateCoupon creates a coupon for a promotion.
func (d *Deps) GenerateCoupon(ctx context.Context, in GenerateCouponInput) (domain.Coupon, error) {
	if in.TenantID == uuid.Nil || in.PromotionID == uuid.Nil {
		return domain.Coupon{}, fmt.Errorf("%w: tenant_id and promotion_id required", domain.ErrInvalidArgument)
	}
	if _, err := d.Promotions.GetByID(ctx, in.TenantID, in.PromotionID); err != nil {
		return domain.Coupon{}, err
	}
	kind := in.Kind
	if kind == "" {
		kind = domain.CouponMulti
	}
	max := in.MaxRedemptions
	if kind == domain.CouponSingle && max == 0 {
		max = 1
	}
	now := d.now()
	c := domain.Coupon{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		PromotionID:    in.PromotionID,
		Code:           strings.ToUpper(strings.TrimSpace(in.Code)),
		Kind:           kind,
		MaxRedemptions: max,
		PrincipalID:    in.PrincipalID,
		StartsAt:       in.StartsAt,
		EndsAt:         in.EndsAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if c.Code == "" {
		c.Code = strings.ToUpper(strings.ReplaceAll(d.newID().String(), "-", "")[:12])
	}
	if err := c.Validate(); err != nil {
		return domain.Coupon{}, err
	}
	if err := d.Coupons.Create(ctx, c); err != nil {
		return domain.Coupon{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCouponGenerated, map[string]any{
		"code": c.Code, "kind": string(c.Kind), "promotionId": c.PromotionID.String(),
	})
	return c, nil
}

// RedeemCouponInput redeems a coupon idempotently.
type RedeemCouponInput struct {
	TenantID       uuid.UUID
	Code           string
	PrincipalID    uuid.UUID
	IdempotencyKey string
	OrderRef       string
	DiscountMinor  int64
	Currency       string
}

// RedeemCoupon redeems a coupon once per idempotency key.
func (d *Deps) RedeemCoupon(ctx context.Context, in RedeemCouponInput) (domain.CouponRedemption, error) {
	if in.TenantID == uuid.Nil || in.Code == "" || in.IdempotencyKey == "" {
		return domain.CouponRedemption{}, fmt.Errorf("%w: tenant_id, code, and idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Coupons.GetRedemptionByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		return existing, nil
	}

	c, err := d.Coupons.GetByCode(ctx, in.TenantID, in.Code)
	if err != nil {
		return domain.CouponRedemption{}, fmt.Errorf("%w: %v", domain.ErrCouponInvalid, err)
	}
	now := d.now()
	if !c.IsValidAt(now) {
		if c.MaxRedemptions > 0 && c.RedeemedCount >= c.MaxRedemptions {
			return domain.CouponRedemption{}, domain.ErrCouponExhausted
		}
		return domain.CouponRedemption{}, domain.ErrCouponInvalid
	}
	if c.Kind == domain.CouponPersonal && c.PrincipalID != nil && in.PrincipalID != *c.PrincipalID {
		return domain.CouponRedemption{}, domain.ErrForbidden
	}
	if c.Kind == domain.CouponSingle && c.RedeemedCount >= 1 {
		return domain.CouponRedemption{}, domain.ErrCouponRedeemed
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "TRY"
	}
	red := domain.CouponRedemption{
		ID:             d.newID(),
		TenantID:       in.TenantID,
		CouponID:       c.ID,
		PrincipalID:    in.PrincipalID,
		IdempotencyKey: in.IdempotencyKey,
		OrderRef:       in.OrderRef,
		DiscountMinor:  in.DiscountMinor,
		Currency:       currency,
		RedeemedAt:     now,
		CreatedAt:      now,
	}
	if err := d.Coupons.CreateRedemption(ctx, red); err != nil {
		if existing, e2 := d.Coupons.GetRedemptionByIdempotency(ctx, in.TenantID, in.IdempotencyKey); e2 == nil {
			return existing, nil
		}
		return domain.CouponRedemption{}, err
	}
	c.RedeemedCount++
	c.UpdatedAt = now
	if err := d.Coupons.Update(ctx, c); err != nil {
		return domain.CouponRedemption{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCouponRedeemed, map[string]any{
		"code": c.Code, "principalId": in.PrincipalID.String(), "orderRef": in.OrderRef,
	})
	return red, nil
}

// ListCoupons returns tenant coupons newest first.
func (d *Deps) ListCoupons(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Coupon, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if d.Coupons == nil {
		return nil, fmt.Errorf("%w: coupon repository not configured", domain.ErrInvalidArgument)
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return d.Coupons.List(ctx, tenantID, limit, offset)
}

// GetCoupon looks up a coupon by code.
func (d *Deps) GetCoupon(ctx context.Context, tenantID uuid.UUID, code string) (domain.Coupon, error) {
	if tenantID == uuid.Nil || strings.TrimSpace(code) == "" {
		return domain.Coupon{}, fmt.Errorf("%w: tenant_id and code required", domain.ErrInvalidArgument)
	}
	return d.Coupons.GetByCode(ctx, tenantID, strings.ToUpper(strings.TrimSpace(code)))
}

// UpdateCouponInput patches coupon eligibility, window, and enablement.
type UpdateCouponInput struct {
	TenantID       uuid.UUID
	Code           string
	Kind           *domain.CouponKind
	MaxRedemptions *int
	StartsAt       *time.Time
	EndsAt         *time.Time
	Enabled        *bool
}

// UpdateCoupon persists coupon configuration changes and writes an audit event.
func (d *Deps) UpdateCoupon(ctx context.Context, in UpdateCouponInput) (domain.Coupon, error) {
	c, err := d.GetCoupon(ctx, in.TenantID, in.Code)
	if err != nil {
		return domain.Coupon{}, err
	}
	now := d.now()
	if in.Kind != nil && *in.Kind != "" {
		c.Kind = *in.Kind
	}
	if in.MaxRedemptions != nil {
		c.MaxRedemptions = *in.MaxRedemptions
	}
	if in.StartsAt != nil {
		c.StartsAt = in.StartsAt
	}
	if in.EndsAt != nil {
		c.EndsAt = in.EndsAt
	}
	if in.Enabled != nil {
		if *in.Enabled {
			if c.EndsAt != nil && !now.Before(*c.EndsAt) {
				c.EndsAt = nil
			}
		} else {
			ended := now
			c.EndsAt = &ended
		}
	}
	c.UpdatedAt = now
	if err := c.Validate(); err != nil {
		return domain.Coupon{}, err
	}
	if err := d.Coupons.Update(ctx, c); err != nil {
		return domain.Coupon{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCouponGenerated, map[string]any{
		"code": c.Code, "action": "update", "enabled": c.IsValidAt(now),
	})
	return c, nil
}
