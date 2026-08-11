package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxThemeLen            = 32
	maxPreferenceLangLen   = 16
)

// Preferences is a 1:1 preference bag for a customer profile.
type Preferences struct {
	ProfileID          uuid.UUID
	FavoriteBrands     []uuid.UUID
	FavoriteCategories []uuid.UUID
	FavoriteProducts   []uuid.UUID
	FavoriteStores     []uuid.UUID
	Delivery           map[string]any
	Payment            map[string]any
	Notification       map[string]any
	Shopping           map[string]any
	Theme              string
	Language           string
	Accessibility      map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Validate checks structural invariants.
func (p Preferences) Validate() error {
	if p.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Theme) > maxThemeLen {
		return fmt.Errorf("%w: theme too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.Language) > maxPreferenceLangLen {
		return fmt.Errorf("%w: language too long", ErrInvalidArgument)
	}
	if err := ensureNoNilUUIDs("favorite_brands", p.FavoriteBrands); err != nil {
		return err
	}
	if err := ensureNoNilUUIDs("favorite_categories", p.FavoriteCategories); err != nil {
		return err
	}
	if err := ensureNoNilUUIDs("favorite_products", p.FavoriteProducts); err != nil {
		return err
	}
	if err := ensureNoNilUUIDs("favorite_stores", p.FavoriteStores); err != nil {
		return err
	}
	return nil
}

func ensureNoNilUUIDs(field string, ids []uuid.UUID) error {
	for i, id := range ids {
		if id == uuid.Nil {
			return fmt.Errorf("%w: %s[%d] is nil", ErrInvalidArgument, field, i)
		}
	}
	return nil
}
