package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// LeaderboardEntry is a stub leaderboard row.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	AccountID   uuid.UUID `json:"accountId"`
	PrincipalID uuid.UUID `json:"principalId"`
	Points      int64     `json:"points"`
	XP          int64     `json:"xp"`
}

// GetLeaderboard returns a stub empty/partial leaderboard.
func (d *Deps) GetLeaderboard(ctx context.Context, tenantID uuid.UUID, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	// Stub: no global ranking store yet — return empty slice.
	_ = ctx
	_ = tenantID
	return []LeaderboardEntry{}, nil
}

// GetAIScores returns stub churn/LTV scores (creates defaults if missing).
func (d *Deps) GetAIScores(ctx context.Context, tenantID, accountID uuid.UUID) (domain.AIScore, error) {
	if existing, err := d.AIScores.Get(ctx, tenantID, accountID); err == nil {
		return existing, nil
	}
	acct, err := d.Accounts.GetAccount(ctx, tenantID, accountID)
	if err != nil {
		return domain.AIScore{}, err
	}
	score := domain.AIScore{
		ID: d.newID(), TenantID: tenantID, AccountID: accountID, PrincipalID: acct.PrincipalID,
		ChurnScore: 0.15, LTVScore: float64(acct.TierPoints) * 0.01, UpdatedAt: d.now(),
	}
	_ = d.AIScores.Upsert(ctx, score)
	return score, nil
}
