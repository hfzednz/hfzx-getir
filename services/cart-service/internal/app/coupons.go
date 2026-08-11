package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// ApplyCouponInput applies a preview coupon code.
type ApplyCouponInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	Code     string
}

// ApplyCoupon attaches a coupon code to the cart (preview; pricing owns discount SoT).
func (d *Deps) ApplyCoupon(ctx context.Context, in ApplyCouponInput) (domain.Cart, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if code == "" {
		return domain.Cart{}, fmt.Errorf("%w: coupon code required", domain.ErrCouponInvalid)
	}
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	if c.HasCoupon(code) {
		return domain.Cart{}, fmt.Errorf("%w: already applied", domain.ErrAlreadyExists)
	}
	now := d.now()
	cp := domain.AppliedCoupon{
		Code:          code,
		DiscountMinor: 0,
		AppliedAt:     now,
		Metadata:      map[string]any{},
	}
	c.Coupons = append(c.Coupons, cp)
	c.Quote = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCouponApplied, map[string]any{
		"code": code,
	})
	return c, nil
}

// RemoveCouponInput removes an applied coupon.
type RemoveCouponInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	Code     string
}

// RemoveCoupon removes a coupon by code.
func (d *Deps) RemoveCoupon(ctx context.Context, in RemoveCouponInput) (domain.Cart, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	idx := -1
	for i, cp := range c.Coupons {
		if strings.EqualFold(cp.Code, code) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.Cart{}, fmt.Errorf("%w", domain.ErrCouponNotApplied)
	}
	c.Coupons = append(c.Coupons[:idx], c.Coupons[idx+1:]...)
	now := d.now()
	c.Quote = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCouponRemoved, map[string]any{
		"code": code,
	})
	return c, nil
}
