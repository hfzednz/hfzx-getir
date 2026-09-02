package httpclients

import (
	"context"
	"strings"

	"github.com/nexora/bff-customer/internal/app/ports"
	"github.com/nexora/bff-customer/internal/domain"
)

// Promo implements ports.PromoClient against promotion-service.
type Promo struct{ Base }

func NewPromo(baseURL string) *Promo { return &Promo{Base: newBase(baseURL)} }

func (c *Promo) ListCoupons(ctx context.Context, tenantID string) ([]map[string]any, error) {
	var raw map[string]any
	if err := c.get(ctx, "/v1/promo/coupons", tenantID, &raw); err != nil {
		return nil, err
	}
	arr, _ := raw["items"].([]any)
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func (c *Promo) GetCoupon(ctx context.Context, tenantID, code string) (map[string]any, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, domain.ErrInvalidArgument
	}
	var out map[string]any
	if err := c.get(ctx, "/v1/promo/coupons/"+code, tenantID, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Promo) EvaluateCoupon(ctx context.Context, tenantID, code string, cartSubtotalMinor int64) (map[string]any, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, domain.ErrInvalidArgument
	}
	qty := 1
	unit := cartSubtotalMinor
	if unit <= 0 {
		unit = 1
	}
	var out map[string]any
	if err := c.post(ctx, "/v1/promo/evaluate", tenantID, map[string]any{
		"currency":    "TRY",
		"couponCodes": []string{code},
		"lines": []map[string]any{
			{"lineId": "basket", "variantId": "basket", "quantity": qty, "unitPriceMinor": unit},
		},
	}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var _ ports.PromoClient = (*Promo)(nil)
