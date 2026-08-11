package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxHouseholdNameLen = 120

// HouseholdMemberRole classifies a member within a household.
type HouseholdMemberRole string

const (
	HouseholdRoleOwner HouseholdMemberRole = "owner"
	HouseholdRoleAdult HouseholdMemberRole = "adult"
	HouseholdRoleChild HouseholdMemberRole = "child"
	HouseholdRoleGuest HouseholdMemberRole = "guest"
)

func (r HouseholdMemberRole) Valid() bool {
	switch r {
	case HouseholdRoleOwner, HouseholdRoleAdult, HouseholdRoleChild, HouseholdRoleGuest:
		return true
	default:
		return false
	}
}

// HouseholdSharingFlags control which resources are shared among members.
type HouseholdSharingFlags struct {
	Addresses bool
	Payments  bool
	Lists     bool
	Wallet    bool
	Loyalty   bool
}

// Household groups related customer profiles for shared delivery/resources.
type Household struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	OwnerProfileID uuid.UUID
	Sharing        HouseholdSharingFlags
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// Validate checks structural invariants.
func (h Household) Validate() error {
	if h.ID == uuid.Nil {
		return fmt.Errorf("%w: household id required", ErrInvalidArgument)
	}
	if h.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if h.OwnerProfileID == uuid.Nil {
		return fmt.Errorf("%w: owner_profile_id required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(h.Name) > maxHouseholdNameLen {
		return fmt.Errorf("%w: household name too long", ErrInvalidArgument)
	}
	return nil
}

// HouseholdMember is a profile membership in a household.
type HouseholdMember struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
	ProfileID   uuid.UUID
	Role        HouseholdMemberRole
	JoinedAt    time.Time
	LeftAt      *time.Time
}

// Validate checks structural invariants.
func (m HouseholdMember) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: household member id required", ErrInvalidArgument)
	}
	if m.HouseholdID == uuid.Nil {
		return fmt.Errorf("%w: household_id required", ErrInvalidArgument)
	}
	if m.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if !m.Role.Valid() {
		return fmt.Errorf("%w: invalid household member role %q", ErrInvalidArgument, m.Role)
	}
	if m.LeftAt != nil && m.LeftAt.Before(m.JoinedAt) {
		return fmt.Errorf("%w: left_at before joined_at", ErrInvariant)
	}
	return nil
}

// IsActive reports whether the membership is current.
func (m HouseholdMember) IsActive() bool {
	return m.LeftAt == nil
}
