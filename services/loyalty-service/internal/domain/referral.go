package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReferralStatus is the invite lifecycle.
type ReferralStatus string

const (
	ReferralOpen      ReferralStatus = "open"
	ReferralApplied   ReferralStatus = "applied"
	ReferralCompleted ReferralStatus = "completed"
	ReferralRejected  ReferralStatus = "rejected"
)

// ReferralCode is owned by a referrer principal.
type ReferralCode struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AccountID   uuid.UUID
	PrincipalID uuid.UUID
	Code        string
	Active      bool
	CreatedAt   time.Time
}

// ReferralEvent tracks apply/complete for a referee.
type ReferralEvent struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	CodeID          uuid.UUID
	ReferrerAccount uuid.UUID
	RefereeAccount  uuid.UUID
	RefereePrincipal uuid.UUID
	Status          ReferralStatus
	OrderID         *uuid.UUID
	PointsGranted   int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks referral event invariants.
func (e ReferralEvent) Validate() error {
	if e.ID == uuid.Nil || e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: referral ids required", ErrInvalidArgument)
	}
	if e.ReferrerAccount == e.RefereeAccount {
		return ErrSelfReferral
	}
	return nil
}
