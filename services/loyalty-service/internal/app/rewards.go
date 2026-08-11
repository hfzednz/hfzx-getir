package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// UnlockRewardInput unlocks a catalog reward for an account.
type UnlockRewardInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
	RewardID  uuid.UUID
}

// UnlockReward creates an unlocked redemption without spending points yet.
func (d *Deps) UnlockReward(ctx context.Context, in UnlockRewardInput) (domain.Redemption, error) {
	reward, err := d.Rewards.GetReward(ctx, in.TenantID, in.RewardID)
	if err != nil {
		return domain.Redemption{}, err
	}
	if !reward.Active {
		return domain.Redemption{}, fmt.Errorf("%w: reward inactive", domain.ErrInvalidArgument)
	}
	if _, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID); err != nil {
		return domain.Redemption{}, err
	}
	now := d.now()
	r := domain.Redemption{
		ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID,
		RewardID: reward.ID, Status: domain.RewardUnlocked, PointsPaid: 0,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Rewards.CreateRedemption(ctx, r); err != nil {
		return domain.Redemption{}, err
	}
	acct, _ := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	d.emit(ctx, acct, domain.EventRewardUnlocked, map[string]any{
		"rewardId": reward.ID.String(), "redemptionId": r.ID.String(),
	})
	return r, nil
}

// RedeemRewardInput spends points to redeem a reward.
type RedeemRewardInput struct {
	TenantID     uuid.UUID
	AccountID    uuid.UUID
	RewardID     uuid.UUID
	RedemptionID uuid.UUID // optional existing unlock
}

// RedeemReward spends points_cost and marks redemption redeemed.
func (d *Deps) RedeemReward(ctx context.Context, in RedeemRewardInput) (domain.Redemption, error) {
	reward, err := d.Rewards.GetReward(ctx, in.TenantID, in.RewardID)
	if err != nil {
		return domain.Redemption{}, err
	}
	if reward.PointsCost > 0 {
		_, _, err := d.RedeemPoints(ctx, RedeemPointsInput{
			TenantID: in.TenantID, AccountID: in.AccountID, Points: reward.PointsCost,
			IdempotencyKey: "reward:" + in.RewardID.String() + ":" + in.AccountID.String(),
			Reference:      "reward:" + reward.Code,
		})
		if err != nil {
			return domain.Redemption{}, err
		}
	}

	now := d.now()
	var r domain.Redemption
	if in.RedemptionID != uuid.Nil {
		r, err = d.Rewards.GetRedemption(ctx, in.TenantID, in.RedemptionID)
		if err != nil {
			return domain.Redemption{}, err
		}
		r.Status = domain.RewardRedeemed
		r.PointsPaid = reward.PointsCost
		r.UpdatedAt = now
		if err := d.Rewards.UpdateRedemption(ctx, r); err != nil {
			return domain.Redemption{}, err
		}
	} else {
		r = domain.Redemption{
			ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID,
			RewardID: reward.ID, Status: domain.RewardRedeemed, PointsPaid: reward.PointsCost,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := d.Rewards.CreateRedemption(ctx, r); err != nil {
			return domain.Redemption{}, err
		}
	}
	acct, _ := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	d.emit(ctx, acct, domain.EventRewardRedeemed, map[string]any{
		"rewardId": reward.ID.String(), "redemptionId": r.ID.String(), "pointsPaid": reward.PointsCost,
	})
	return r, nil
}
