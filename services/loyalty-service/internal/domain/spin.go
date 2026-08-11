package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SpinPrize is a weighted prize in a campaign.
type SpinPrize struct {
	Code   string
	Weight int
	Points int64
}

// SpinCampaign is a weighted prize wheel.
type SpinCampaign struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Title     string
	Prizes    []SpinPrize
	Active    bool
	CreatedAt time.Time
}

// TotalWeight sums prize weights.
func (c SpinCampaign) TotalWeight() int {
	sum := 0
	for _, p := range c.Prizes {
		if p.Weight > 0 {
			sum += p.Weight
		}
	}
	return sum
}

// PickPrize selects a prize given a roll in [0, totalWeight).
func (c SpinCampaign) PickPrize(roll int) (SpinPrize, error) {
	total := c.TotalWeight()
	if total <= 0 {
		return SpinPrize{}, fmt.Errorf("%w: no weighted prizes", ErrInvariant)
	}
	if roll < 0 {
		roll = 0
	}
	roll = roll % total
	cursor := 0
	for _, p := range c.Prizes {
		if p.Weight <= 0 {
			continue
		}
		cursor += p.Weight
		if roll < cursor {
			return p, nil
		}
	}
	return c.Prizes[len(c.Prizes)-1], nil
}

// SpinResult is a recorded spin.
type SpinResult struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AccountID  uuid.UUID
	CampaignID uuid.UUID
	PrizeCode  string
	PointsWon  int64
	CreatedAt  time.Time
}
