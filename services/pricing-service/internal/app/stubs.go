package app

import (
	"context"

	"github.com/nexora/pricing-service/internal/app/ports"
)

// NoopPromoClient returns zero discounts (default when promotion-service unreachable).
type NoopPromoClient struct{}

func (NoopPromoClient) Evaluate(_ context.Context, _ ports.PromoEvaluateRequest) (ports.PromoEvaluateResult, error) {
	return ports.PromoEvaluateResult{}, nil
}

// NoopHintClient returns empty inventory hints.
type NoopHintClient struct{}

func (NoopHintClient) Hint(_ context.Context, _ ports.DynamicHintRequest) (ports.DynamicHintResult, error) {
	return ports.DynamicHintResult{}, nil
}
