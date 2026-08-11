package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// GetAIModel returns the stored AI customer model.
func (d *Deps) GetAIModel(ctx context.Context, profileID uuid.UUID) (domain.AICustomerModel, error) {
	if profileID == uuid.Nil {
		return domain.AICustomerModel{}, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	return d.AIModels.Get(ctx, profileID)
}

// RecomputeAIModel applies a simple heuristic from activity/orders summary (no ML).
func (d *Deps) RecomputeAIModel(ctx context.Context, profileID uuid.UUID, summary domain.ActivityOrdersSummary) (domain.AICustomerModel, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return domain.AICustomerModel{}, err
	}

	churn := clamp01(float64(summary.DaysSinceLast) / 90.0)
	if summary.CancelRate > 0 {
		churn = clamp01(churn + summary.CancelRate*0.3)
	}
	freq := float64(summary.OrderCount30d)
	avgMinor := int64(summary.AvgOrderValue * 100)
	if avgMinor < 0 {
		avgMinor = 0
	}
	priceSens := clamp01(1.0 - float64(summary.OrderCount90d)/20.0)

	cats := map[string]any{}
	for i, c := range summary.PreferredCats {
		cats[c] = 1.0 - float64(i)*0.1
	}

	now := d.now()
	m := domain.AICustomerModel{
		ProfileID:           profileID,
		Frequency:           freq,
		AvgOrderValueMinor:  avgMinor,
		ChurnProb:           churn,
		PriceSensitivity:    priceSens,
		CategoryAffinity:    cats,
		ModelVersion:        "heuristic-v1",
		UpdatedAt:           now,
		CreatedAt:           now,
	}
	if existing, err := d.AIModels.Get(ctx, profileID); err == nil {
		m.CreatedAt = existing.CreatedAt
		m.PreferredOrderHours = existing.PreferredOrderHours
		m.PreferredOrderWeekdays = existing.PreferredOrderWeekdays
		m.BrandAffinity = existing.BrandAffinity
	}
	if err := m.Validate(); err != nil {
		return domain.AICustomerModel{}, err
	}
	if err := d.AIModels.Upsert(ctx, m); err != nil {
		return domain.AICustomerModel{}, err
	}
	return m, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
