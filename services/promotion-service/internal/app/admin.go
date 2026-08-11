package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// SimulateCart runs Evaluate and persists the simulation for admin review.
func (d *Deps) SimulateCart(ctx context.Context, eval domain.EvaluateContext) (domain.EvaluateResult, domain.Simulation, error) {
	res, err := d.EvaluateCart(ctx, eval)
	if err != nil {
		return domain.EvaluateResult{}, domain.Simulation{}, err
	}
	now := d.now()
	sim := domain.Simulation{
		ID:       d.newID(),
		TenantID: eval.TenantID,
		RequestPayload: map[string]any{
			"currency":      eval.Currency,
			"principalId":   eval.PrincipalID.String(),
			"couponCodes":   eval.CouponCodes,
			"shippingMinor": eval.ShippingMinor,
			"segmentIds":    eval.SegmentIDs,
			"lineCount":     len(eval.Lines),
		},
		ResultPayload: map[string]any{
			"totalDiscountMinor":    res.TotalDiscountMinor,
			"shippingDiscountMinor": res.ShippingDiscountMinor,
			"discountCount":         len(res.Discounts),
			"currency":              res.Currency,
		},
		CreatedAt: now,
	}
	if d.Simulations != nil {
		if err := d.Simulations.Create(ctx, sim); err != nil {
			return res, sim, err
		}
	}
	return res, sim, nil
}

// ListSimulations returns recent simulations for admin.
func (d *Deps) ListSimulations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Simulation, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Simulations.List(ctx, tenantID, limit, offset)
}

// PublishPendingOutbox drains pending outbox messages (stub publish).
func (d *Deps) PublishPendingOutbox(ctx context.Context, limit int) (int, error) {
	if d.Outbox == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	pending, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	now := d.now()
	for _, msg := range pending {
		if d.Publisher != nil {
			if err := d.Publisher.Publish(ctx, msg.Topic, msg.Key, msg.Payload); err != nil {
				return n, err
			}
		}
		if err := d.Outbox.MarkPublished(ctx, msg.ID, now); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// AdminListOverview returns campaign/promo counts for admin dashboards.
func (d *Deps) AdminListOverview(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	camps, err := d.Campaigns.List(ctx, tenantID, nil, 1000, 0)
	if err != nil {
		return nil, err
	}
	active := 0
	for _, c := range camps {
		if c.Status == domain.CampaignActive {
			active++
		}
	}
	return map[string]any{
		"campaignCount":       len(camps),
		"activeCampaignCount": active,
	}, nil
}
