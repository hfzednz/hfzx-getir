package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Personalization holds homepage/category/recommendation/search/delivery/promotion prefs
// and observed shopping habits for a profile.
type Personalization struct {
	ProfileID       uuid.UUID
	Homepage        map[string]any
	Category        map[string]any
	Recommendation  map[string]any
	Search          map[string]any
	Delivery        map[string]any
	Promotion       map[string]any
	ShoppingHabits  map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks structural invariants.
func (p Personalization) Validate() error {
	if p.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	return nil
}
