package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// SpinInput spins a campaign wheel.
type SpinInput struct {
	TenantID   uuid.UUID
	AccountID  uuid.UUID
	CampaignID uuid.UUID
}

// Spin picks a weighted prize using injectable Rand and grants points.
func (d *Deps) Spin(ctx context.Context, in SpinInput) (domain.SpinResult, error) {
	campaign, err := d.Spins.GetCampaign(ctx, in.TenantID, in.CampaignID)
	if err != nil {
		return domain.SpinResult{}, err
	}
	if !campaign.Active {
		return domain.SpinResult{}, fmt.Errorf("%w: campaign inactive", domain.ErrInvalidArgument)
	}
	acct, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	if err != nil {
		return domain.SpinResult{}, err
	}
	total := campaign.TotalWeight()
	if total <= 0 {
		return domain.SpinResult{}, fmt.Errorf("%w: no prizes", domain.ErrInvariant)
	}
	roll := d.intn(total)
	prize, err := campaign.PickPrize(roll)
	if err != nil {
		return domain.SpinResult{}, err
	}
	now := d.now()
	res := domain.SpinResult{
		ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID, CampaignID: campaign.ID,
		PrizeCode: prize.Code, PointsWon: prize.Points, CreatedAt: now,
	}
	if err := d.Spins.CreateSpin(ctx, res); err != nil {
		return domain.SpinResult{}, err
	}
	if prize.Points > 0 {
		_, _, err = d.grantPoints(ctx, acct, prize.Points, domain.PointGrant,
			"spin:"+prize.Code, "spin:"+res.ID.String(), map[string]any{"roll": roll})
		if err != nil {
			return domain.SpinResult{}, err
		}
	}
	return res, nil
}
