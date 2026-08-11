package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CampaignStatus is the campaign lifecycle state.
type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignScheduled CampaignStatus = "scheduled"
	CampaignActive    CampaignStatus = "active"
	CampaignPaused    CampaignStatus = "paused"
	CampaignExpired   CampaignStatus = "expired"
)

// Valid reports whether the status is recognized.
func (s CampaignStatus) Valid() bool {
	switch s {
	case CampaignDraft, CampaignScheduled, CampaignActive, CampaignPaused, CampaignExpired:
		return true
	default:
		return false
	}
}

// Campaign groups promotions under a schedule and status.
type Campaign struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Status      CampaignStatus
	StartsAt    *time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}

// Validate checks campaign invariants.
func (c Campaign) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: campaign id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if c.Name == "" {
		return fmt.Errorf("%w: campaign name required", ErrInvalidArgument)
	}
	if !c.Status.Valid() {
		return fmt.Errorf("%w: invalid campaign status %q", ErrInvalidArgument, c.Status)
	}
	if c.StartsAt != nil && c.EndsAt != nil && !c.EndsAt.After(*c.StartsAt) {
		return fmt.Errorf("%w: ends_at must be after starts_at", ErrInvalidArgument)
	}
	return nil
}

// IsActiveAt reports whether the campaign is usable at the given time.
func (c Campaign) IsActiveAt(now time.Time) bool {
	if c.Status != CampaignActive {
		return false
	}
	if c.StartsAt != nil && now.Before(*c.StartsAt) {
		return false
	}
	if c.EndsAt != nil && !now.Before(*c.EndsAt) {
		return false
	}
	return true
}

// CanTransition reports whether status change from → to is allowed.
func (c Campaign) CanTransition(to CampaignStatus) bool {
	switch c.Status {
	case CampaignDraft:
		return to == CampaignScheduled || to == CampaignActive || to == CampaignExpired
	case CampaignScheduled:
		return to == CampaignActive || to == CampaignPaused || to == CampaignExpired
	case CampaignActive:
		return to == CampaignPaused || to == CampaignExpired
	case CampaignPaused:
		return to == CampaignActive || to == CampaignExpired
	case CampaignExpired:
		return false
	default:
		return false
	}
}
