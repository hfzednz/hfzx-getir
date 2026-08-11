package memory

import (
	"context"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
)

// PromoClient is a configurable mock PromoClient for tests.
type PromoClient struct {
	Calls          atomic.Int64
	Fail           bool
	DiscountMinor  int64
	Discounts      []ports.PromoDiscountResult
	Err            error
}

func (c *PromoClient) Evaluate(_ context.Context, req ports.PromoEvaluateRequest) (ports.PromoEvaluateResult, error) {
	c.Calls.Add(1)
	if c.Fail {
		return ports.PromoEvaluateResult{}, c.Err
	}
	if c.Err != nil {
		return ports.PromoEvaluateResult{}, c.Err
	}
	if len(c.Discounts) > 0 {
		var total int64
		for _, d := range c.Discounts {
			total += d.DiscountMinor
		}
		return ports.PromoEvaluateResult{Discounts: c.Discounts, DiscountMinor: total}, nil
	}
	if c.DiscountMinor > 0 {
		return ports.PromoEvaluateResult{
			DiscountMinor: c.DiscountMinor,
			Discounts: []ports.PromoDiscountResult{{
				PromotionID: "promo-mock", Code: "MOCK10",
				DiscountMinor: c.DiscountMinor, Description: "mock discount",
			}},
		}, nil
	}
	_ = req
	return ports.PromoEvaluateResult{}, nil
}

var _ ports.PromoClient = (*PromoClient)(nil)

// HintClient is a configurable DynamicHintClient stub.
type HintClient struct {
	Calls        atomic.Int64
	AvailableQty *int
	ByVariant    map[uuid.UUID]int
}

func (c *HintClient) Hint(_ context.Context, req ports.DynamicHintRequest) (ports.DynamicHintResult, error) {
	c.Calls.Add(1)
	if c.ByVariant != nil {
		if q, ok := c.ByVariant[req.VariantID]; ok {
			qq := q
			return ports.DynamicHintResult{AvailableQty: &qq}, nil
		}
	}
	return ports.DynamicHintResult{AvailableQty: c.AvailableQty}, nil
}

var _ ports.DynamicHintClient = (*HintClient)(nil)
