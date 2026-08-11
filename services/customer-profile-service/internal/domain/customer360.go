package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Customer360 is an aggregate CRM view assembled by the application layer.
type Customer360 struct {
	Profile         CustomerProfile
	Preferences     *Preferences
	Addresses       []Address
	Tags            []ProfileTag
	Consents        []Consent
	Notes           []CRMNote
	Timeline        []TimelineEvent
	Segments        []SegmentMembership
	Personalization *Personalization
	AIModel         *AICustomerModel
	Household       *Household
}

// Ensure Customer360 stays a pure DTO (no persistence invariants beyond nested types).
func (v Customer360) ProfileID() uuid.UUID {
	return v.Profile.ID
}

// StampUpdated is a helper for adapters that need a consistent clock on nested views.
func StampUpdated(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
}

// ValidateProfileID is a tiny shared guard used by application helpers.
func ValidateProfileID(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	return nil
}
