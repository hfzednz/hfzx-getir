package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RewardStatus is the unlock/redeem lifecycle.
type RewardStatus string

const (
	RewardAvailable RewardStatus = "available"
	RewardUnlocked  RewardStatus = "unlocked"
	RewardRedeemed  RewardStatus = "redeemed"
)

// Reward is a catalog reward definition.
type Reward struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Code          string
	Title         string
	PointsCost    int64
	Active        bool
	Metadata      map[string]any
	CreatedAt     time.Time
}

// Redemption is an account's unlocked/redeemed reward instance.
type Redemption struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AccountID  uuid.UUID
	RewardID   uuid.UUID
	Status     RewardStatus
	PointsPaid int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks redemption invariants.
func (r Redemption) Validate() error {
	if r.ID == uuid.Nil || r.TenantID == uuid.Nil || r.AccountID == uuid.Nil || r.RewardID == uuid.Nil {
		return fmt.Errorf("%w: redemption ids required", ErrInvalidArgument)
	}
	switch r.Status {
	case RewardAvailable, RewardUnlocked, RewardRedeemed:
	default:
		return fmt.Errorf("%w: invalid reward status", ErrInvalidArgument)
	}
	return nil
}
