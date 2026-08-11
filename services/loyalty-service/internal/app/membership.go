package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// EvaluateMembershipInput evaluates tier upgrade for an account.
type EvaluateMembershipInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
}

// EvaluateMembership upgrades membership when tier_points cross thresholds.
func (d *Deps) EvaluateMembership(ctx context.Context, in EvaluateMembershipInput) (domain.Membership, bool, error) {
	acct, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	if err != nil {
		return domain.Membership{}, false, err
	}
	tiers, err := d.Memberships.ListTiers(ctx, in.TenantID)
	if err != nil || len(tiers) == 0 {
		tiers = domain.DefaultTiers()
	}
	best := domain.BestTierForPoints(tiers, acct.TierPoints)

	cur, err := d.Memberships.GetMembership(ctx, in.TenantID, in.AccountID)
	now := d.now()
	if err != nil {
		m := domain.Membership{
			ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID,
			Tier: best.Code, Since: now, UpdatedAt: now,
		}
		if err := d.Memberships.UpsertMembership(ctx, m); err != nil {
			return domain.Membership{}, false, err
		}
		return m, true, nil
	}

	upgraded := false
	if cur.Tier != best.Code {
		prevRank := tierRank(tiers, cur.Tier)
		if best.Rank > prevRank {
			upgraded = true
			d.emit(ctx, acct, domain.EventMembershipUpgraded, map[string]any{
				"from": string(cur.Tier), "to": string(best.Code), "tierPoints": acct.TierPoints,
			})
		} else if best.Rank < prevRank {
			d.emit(ctx, acct, domain.EventMembershipDowngraded, map[string]any{
				"from": string(cur.Tier), "to": string(best.Code), "tierPoints": acct.TierPoints,
			})
		}
		cur.Tier = best.Code
		if upgraded {
			cur.Since = now
		}
		cur.UpdatedAt = now
		if err := d.Memberships.UpsertMembership(ctx, cur); err != nil {
			return domain.Membership{}, false, err
		}
	}
	return cur, upgraded, nil
}

// GetMembership returns current membership.
func (d *Deps) GetMembership(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Membership, error) {
	return d.Memberships.GetMembership(ctx, tenantID, accountID)
}

func tierRank(tiers []domain.TierConfig, code domain.TierCode) int {
	for _, t := range tiers {
		if t.Code == code {
			return t.Rank
		}
	}
	return 0
}

// EnsureMembership is a thin alias used by HTTP.
func (d *Deps) EnsureMembership(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Membership, error) {
	m, _, err := d.EvaluateMembership(ctx, EvaluateMembershipInput{TenantID: tenantID, AccountID: accountID})
	if err != nil {
		return domain.Membership{}, fmt.Errorf("%w", err)
	}
	return m, nil
}
