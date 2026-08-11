package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ProfileStatus is the lifecycle state of a customer profile.
type ProfileStatus string

const (
	ProfileStatusActive  ProfileStatus = "active"
	ProfileStatusMerged  ProfileStatus = "merged"
	ProfileStatusDeleted ProfileStatus = "deleted"
)

func (s ProfileStatus) Valid() bool {
	switch s {
	case ProfileStatusActive, ProfileStatusMerged, ProfileStatusDeleted:
		return true
	default:
		return false
	}
}

// Gender is an optional demographic attribute.
type Gender string

const (
	GenderUnspecified   Gender = "unspecified"
	GenderFemale        Gender = "female"
	GenderMale          Gender = "male"
	GenderNonBinary     Gender = "non_binary"
	GenderOther         Gender = "other"
	GenderPreferNotSay  Gender = "prefer_not_to_say"
)

func (g Gender) Valid() bool {
	switch g {
	case GenderUnspecified, GenderFemale, GenderMale, GenderNonBinary, GenderOther, GenderPreferNotSay:
		return true
	default:
		return false
	}
}

const (
	maxDisplayNameLen = 120
	maxFullNameLen    = 200
	maxNicknameLen    = 80
	maxAvatarURLLen   = 2048
	maxLanguageLen    = 16
	maxCityLen        = 120
	maxTimezoneLen    = 64
	maxOccupationLen  = 120
	maxFamilySize     = 50
)

// CustomerProfile is the aggregate root for profile & CRM attributes.
// Keyed by PrincipalID from identity-service — no credentials live here.
type CustomerProfile struct {
	ID           uuid.UUID
	PrincipalID  uuid.UUID
	TenantID     uuid.UUID
	DisplayName  string
	FullName     string
	Nickname     string
	AvatarURL    string
	Gender       Gender
	Birthday     *time.Time
	Language     string
	CountryCode  string
	City         string
	Timezone     string
	Occupation   string
	FamilySize   int
	Dietary      map[string]any
	Accessibility map[string]any
	Status       ProfileStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// Validate checks structural invariants.
func (p CustomerProfile) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: profile id required", ErrInvalidArgument)
	}
	if p.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !p.Status.Valid() {
		return fmt.Errorf("%w: invalid profile status %q", ErrInvalidArgument, p.Status)
	}
	if !p.Gender.Valid() {
		return fmt.Errorf("%w: invalid gender %q", ErrInvalidArgument, p.Gender)
	}
	if utf8.RuneCountInString(p.DisplayName) > maxDisplayNameLen {
		return fmt.Errorf("%w: display_name too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.FullName) > maxFullNameLen {
		return fmt.Errorf("%w: full_name too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Nickname) > maxNicknameLen {
		return fmt.Errorf("%w: nickname too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.AvatarURL) > maxAvatarURLLen {
		return fmt.Errorf("%w: avatar_url too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Language) > maxLanguageLen {
		return fmt.Errorf("%w: language too long", ErrInvalidArgument)
	}
	if cc := p.CountryCode; cc != "" && len(cc) != 2 {
		return fmt.Errorf("%w: country_code must be ISO-3166 alpha-2", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.City) > maxCityLen {
		return fmt.Errorf("%w: city too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Timezone) > maxTimezoneLen {
		return fmt.Errorf("%w: timezone too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Occupation) > maxOccupationLen {
		return fmt.Errorf("%w: occupation too long", ErrInvalidArgument)
	}
	if p.FamilySize < 0 || p.FamilySize > maxFamilySize {
		return fmt.Errorf("%w: family_size out of range", ErrInvalidArgument)
	}
	if p.Birthday != nil {
		if p.Birthday.After(time.Now().UTC()) {
			return fmt.Errorf("%w: birthday cannot be in the future", ErrInvalidArgument)
		}
	}
	if p.Status == ProfileStatusDeleted && p.DeletedAt == nil {
		return fmt.Errorf("%w: deleted profile requires deleted_at", ErrInvariant)
	}
	if p.DeletedAt != nil && p.Status != ProfileStatusDeleted {
		return fmt.Errorf("%w: deleted_at set but status is %s", ErrInvariant, p.Status)
	}
	return nil
}

// IsActive reports whether the profile may be mutated by normal flows.
func (p CustomerProfile) IsActive() bool {
	return p.Status == ProfileStatusActive && p.DeletedAt == nil
}
