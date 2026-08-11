package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MissionStatus is progress lifecycle.
type MissionStatus string

const (
	MissionActive    MissionStatus = "active"
	MissionCompleted MissionStatus = "completed"
	MissionExpired   MissionStatus = "expired"
)

// Mission is a challenge definition.
type Mission struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Code         string
	Title        string
	TargetCount  int64
	RewardPoints int64
	Achievement  string // optional achievement code on complete
	Active       bool
	CreatedAt    time.Time
}

// MissionProgress tracks an account against a mission.
type MissionProgress struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AccountID  uuid.UUID
	MissionID  uuid.UUID
	Progress   int64
	Status     MissionStatus
	UpdatedAt  time.Time
	CompletedAt *time.Time
}

// Validate checks progress invariants.
func (p MissionProgress) Validate() error {
	if p.ID == uuid.Nil || p.TenantID == uuid.Nil || p.AccountID == uuid.Nil || p.MissionID == uuid.Nil {
		return fmt.Errorf("%w: mission progress ids required", ErrInvalidArgument)
	}
	if p.Progress < 0 {
		return fmt.Errorf("%w: progress must be non-negative", ErrInvariant)
	}
	return nil
}
