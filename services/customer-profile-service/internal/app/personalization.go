package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// UpdatePersonalizationInput upserts personalization maps.
type UpdatePersonalizationInput struct {
	ProfileID      uuid.UUID
	Homepage       map[string]any
	Category       map[string]any
	Recommendation map[string]any
	Search         map[string]any
	Delivery       map[string]any
	Promotion      map[string]any
	ShoppingHabits map[string]any
}

// GetPersonalization returns the personalization profile (empty defaults if missing).
func (d *Deps) GetPersonalization(ctx context.Context, profileID uuid.UUID) (domain.Personalization, error) {
	if profileID == uuid.Nil {
		return domain.Personalization{}, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return domain.Personalization{}, err
	}
	pers, err := d.Personalization.Get(ctx, profileID)
	if err == domain.ErrNotFound {
		return domain.Personalization{ProfileID: profileID}, nil
	}
	return pers, err
}

// UpdatePersonalization upserts personalization data.
func (d *Deps) UpdatePersonalization(ctx context.Context, in UpdatePersonalizationInput) (domain.Personalization, error) {
	if _, err := d.requireActiveProfile(ctx, in.ProfileID); err != nil {
		return domain.Personalization{}, err
	}
	pers, err := d.Personalization.Get(ctx, in.ProfileID)
	now := d.now()
	if err != nil && err != domain.ErrNotFound {
		return domain.Personalization{}, err
	}
	if err == domain.ErrNotFound {
		pers = domain.Personalization{ProfileID: in.ProfileID, CreatedAt: now}
	}
	if in.Homepage != nil {
		pers.Homepage = in.Homepage
	}
	if in.Category != nil {
		pers.Category = in.Category
	}
	if in.Recommendation != nil {
		pers.Recommendation = in.Recommendation
	}
	if in.Search != nil {
		pers.Search = in.Search
	}
	if in.Delivery != nil {
		pers.Delivery = in.Delivery
	}
	if in.Promotion != nil {
		pers.Promotion = in.Promotion
	}
	if in.ShoppingHabits != nil {
		pers.ShoppingHabits = in.ShoppingHabits
	}
	pers.UpdatedAt = now
	if err := pers.Validate(); err != nil {
		return domain.Personalization{}, err
	}
	if err := d.Personalization.Upsert(ctx, pers); err != nil {
		return domain.Personalization{}, err
	}
	return pers, nil
}
