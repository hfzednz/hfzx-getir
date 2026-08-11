package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// UpsertPreferencesInput replaces preference settings.
type UpsertPreferencesInput struct {
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
	TraceID            string
}

// GetPreferences returns preferences for a profile (empty defaults if missing).
func (d *Deps) GetPreferences(ctx context.Context, profileID uuid.UUID) (domain.Preferences, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return domain.Preferences{}, err
	}
	prefs, err := d.Preferences.Get(ctx, profileID)
	if err == domain.ErrNotFound {
		return domain.Preferences{ProfileID: profileID}, nil
	}
	return prefs, err
}

// UpsertPreferences creates or updates preferences and emits PreferenceChanged.
func (d *Deps) UpsertPreferences(ctx context.Context, in UpsertPreferencesInput) (domain.Preferences, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.Preferences{}, err
	}
	prefs, err := d.Preferences.Get(ctx, in.ProfileID)
	now := d.now()
	if err != nil && err != domain.ErrNotFound {
		return domain.Preferences{}, err
	}
	if err == domain.ErrNotFound {
		prefs = domain.Preferences{ProfileID: in.ProfileID, CreatedAt: now}
	}
	if in.FavoriteBrands != nil {
		prefs.FavoriteBrands = in.FavoriteBrands
	}
	if in.FavoriteCategories != nil {
		prefs.FavoriteCategories = in.FavoriteCategories
	}
	if in.FavoriteProducts != nil {
		prefs.FavoriteProducts = in.FavoriteProducts
	}
	if in.FavoriteStores != nil {
		prefs.FavoriteStores = in.FavoriteStores
	}
	if in.Delivery != nil {
		prefs.Delivery = in.Delivery
	}
	if in.Payment != nil {
		prefs.Payment = in.Payment
	}
	if in.Notification != nil {
		prefs.Notification = in.Notification
	}
	if in.Shopping != nil {
		prefs.Shopping = in.Shopping
	}
	if in.Theme != "" {
		prefs.Theme = in.Theme
	}
	if in.Language != "" {
		prefs.Language = in.Language
	}
	if in.Accessibility != nil {
		prefs.Accessibility = in.Accessibility
	}
	prefs.UpdatedAt = now
	if err := prefs.Validate(); err != nil {
		return domain.Preferences{}, err
	}
	if err := d.Preferences.Upsert(ctx, prefs); err != nil {
		return domain.Preferences{}, err
	}
	d.publish(ctx, ports.TopicPreferenceEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventPreferenceChanged,
		"occurredAt": d.now(), "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "traceId": in.TraceID,
	})
	return prefs, nil
}
