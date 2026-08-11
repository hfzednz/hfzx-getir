package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TierCode identifies a membership tier.
type TierCode string

const (
	TierStandard   TierCode = "standard"
	TierSilver     TierCode = "silver"
	TierGold       TierCode = "gold"
	TierPlatinum   TierCode = "platinum"
	TierDiamond    TierCode = "diamond"
	TierVIP        TierCode = "vip"
	TierCorporate  TierCode = "corporate"
	TierFamily     TierCode = "family"
	TierStudent    TierCode = "student"
	TierCustom     TierCode = "custom"
)

// TierConfig defines upgrade thresholds (tier_points).
type TierConfig struct {
	Code              TierCode
	Name              string
	ThresholdPoints   int64
	Rank              int
	Benefits          map[string]any
}

// DefaultTiers returns the standard progressive tier ladder.
func DefaultTiers() []TierConfig {
	return []TierConfig{
		{Code: TierStandard, Name: "Standard", ThresholdPoints: 0, Rank: 0},
		{Code: TierSilver, Name: "Silver", ThresholdPoints: 1000, Rank: 1},
		{Code: TierGold, Name: "Gold", ThresholdPoints: 5000, Rank: 2},
		{Code: TierPlatinum, Name: "Platinum", ThresholdPoints: 15000, Rank: 3},
		{Code: TierDiamond, Name: "Diamond", ThresholdPoints: 50000, Rank: 4},
		{Code: TierVIP, Name: "VIP", ThresholdPoints: 100000, Rank: 5},
	}
}

// BestTierForPoints picks the highest tier whose threshold is met.
func BestTierForPoints(tiers []TierConfig, tierPoints int64) TierConfig {
	best := TierConfig{Code: TierStandard, Name: "Standard", ThresholdPoints: 0, Rank: 0}
	for _, t := range tiers {
		if tierPoints >= t.ThresholdPoints && t.Rank >= best.Rank {
			best = t
		}
	}
	return best
}

// Membership is the current tier assignment for an account.
type Membership struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AccountID  uuid.UUID
	Tier       TierCode
	Since      time.Time
	UpdatedAt  time.Time
}

// Validate checks membership invariants.
func (m Membership) Validate() error {
	if m.ID == uuid.Nil || m.TenantID == uuid.Nil || m.AccountID == uuid.Nil {
		return fmt.Errorf("%w: membership ids required", ErrInvalidArgument)
	}
	if m.Tier == "" {
		return fmt.Errorf("%w: tier required", ErrInvalidArgument)
	}
	return nil
}
