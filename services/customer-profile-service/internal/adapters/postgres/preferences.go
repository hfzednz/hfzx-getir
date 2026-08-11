package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// PreferencesRepo persists preferences.
type PreferencesRepo struct{ DB *sql.DB }

var _ ports.PreferencesRepository = (*PreferencesRepo)(nil)

func (r *PreferencesRepo) Get(ctx context.Context, profileID uuid.UUID) (domain.Preferences, error) {
	var p domain.Preferences
	var brands, cats, products, stores UUIDArray
	var delivery, payment, notification, shopping, accessibility JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT profile_id, favorite_brands, favorite_categories, favorite_products, favorite_stores,
			delivery, payment, notification, shopping, theme, language, accessibility, created_at, updated_at
		FROM preferences WHERE profile_id=$1`, profileID).Scan(
		&p.ProfileID, &brands, &cats, &products, &stores,
		&delivery, &payment, &notification, &shopping, &p.Theme, &p.Language, &accessibility, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.Preferences{}, mapNotFound(err)
	}
	p.FavoriteBrands = []uuid.UUID(brands)
	p.FavoriteCategories = []uuid.UUID(cats)
	p.FavoriteProducts = []uuid.UUID(products)
	p.FavoriteStores = []uuid.UUID(stores)
	p.Delivery = map[string]any(delivery)
	p.Payment = map[string]any(payment)
	p.Notification = map[string]any(notification)
	p.Shopping = map[string]any(shopping)
	p.Accessibility = map[string]any(accessibility)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *PreferencesRepo) Upsert(ctx context.Context, p domain.Preferences) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO preferences (
			profile_id, favorite_brands, favorite_categories, favorite_products, favorite_stores,
			delivery, payment, notification, shopping, theme, language, accessibility, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8,$9,$10,$11,$12,$13,$14
		)
		ON CONFLICT (profile_id) DO UPDATE SET
			favorite_brands=EXCLUDED.favorite_brands,
			favorite_categories=EXCLUDED.favorite_categories,
			favorite_products=EXCLUDED.favorite_products,
			favorite_stores=EXCLUDED.favorite_stores,
			delivery=EXCLUDED.delivery,
			payment=EXCLUDED.payment,
			notification=EXCLUDED.notification,
			shopping=EXCLUDED.shopping,
			theme=EXCLUDED.theme,
			language=EXCLUDED.language,
			accessibility=EXCLUDED.accessibility,
			updated_at=EXCLUDED.updated_at`,
		p.ProfileID, UUIDArray(p.FavoriteBrands), UUIDArray(p.FavoriteCategories),
		UUIDArray(p.FavoriteProducts), UUIDArray(p.FavoriteStores),
		JSONMap(metaGetMap(p.Delivery)), JSONMap(metaGetMap(p.Payment)),
		JSONMap(metaGetMap(p.Notification)), JSONMap(metaGetMap(p.Shopping)),
		p.Theme, p.Language, JSONMap(metaGetMap(p.Accessibility)), p.CreatedAt, p.UpdatedAt,
	)
	return err
}
