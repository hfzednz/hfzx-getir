package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// PersonalizationRepo persists personalization profiles.
type PersonalizationRepo struct{ DB *sql.DB }

var _ ports.PersonalizationRepository = (*PersonalizationRepo)(nil)

func (r *PersonalizationRepo) Get(ctx context.Context, profileID uuid.UUID) (domain.Personalization, error) {
	var p domain.Personalization
	var homepage, category, recommendation, search, delivery, promotion, habits JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT profile_id, homepage, category, recommendation, search, delivery, promotion, shopping_habits,
			created_at, updated_at
		FROM personalization WHERE profile_id=$1`, profileID).Scan(
		&p.ProfileID, &homepage, &category, &recommendation, &search, &delivery, &promotion, &habits,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.Personalization{}, mapNotFound(err)
	}
	p.Homepage = map[string]any(homepage)
	p.Category = map[string]any(category)
	p.Recommendation = map[string]any(recommendation)
	p.Search = map[string]any(search)
	p.Delivery = map[string]any(delivery)
	p.Promotion = map[string]any(promotion)
	p.ShoppingHabits = map[string]any(habits)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *PersonalizationRepo) Upsert(ctx context.Context, p domain.Personalization) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO personalization (
			profile_id, homepage, category, recommendation, search, delivery, promotion, shopping_habits,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (profile_id) DO UPDATE SET
			homepage=EXCLUDED.homepage, category=EXCLUDED.category, recommendation=EXCLUDED.recommendation,
			search=EXCLUDED.search, delivery=EXCLUDED.delivery, promotion=EXCLUDED.promotion,
			shopping_habits=EXCLUDED.shopping_habits, updated_at=EXCLUDED.updated_at`,
		p.ProfileID, JSONMap(metaGetMap(p.Homepage)), JSONMap(metaGetMap(p.Category)),
		JSONMap(metaGetMap(p.Recommendation)), JSONMap(metaGetMap(p.Search)), JSONMap(metaGetMap(p.Delivery)),
		JSONMap(metaGetMap(p.Promotion)), JSONMap(metaGetMap(p.ShoppingHabits)), p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// AIModelRepo persists AI customer model scores.
type AIModelRepo struct{ DB *sql.DB }

var _ ports.AIModelRepository = (*AIModelRepo)(nil)

func (r *AIModelRepo) Get(ctx context.Context, profileID uuid.UUID) (domain.AICustomerModel, error) {
	var m domain.AICustomerModel
	var hours, weekdays IntArray
	var brand, category JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT profile_id, frequency, avg_order_value_minor, churn_prob, preferred_order_hours,
			preferred_order_weekdays, price_sensitivity, brand_affinity, category_affinity,
			model_version, updated_at, created_at
		FROM ai_customer_models WHERE profile_id=$1`, profileID).Scan(
		&m.ProfileID, &m.Frequency, &m.AvgOrderValueMinor, &m.ChurnProb, &hours,
		&weekdays, &m.PriceSensitivity, &brand, &category,
		&m.ModelVersion, &m.UpdatedAt, &m.CreatedAt,
	)
	if err != nil {
		return domain.AICustomerModel{}, mapNotFound(err)
	}
	m.PreferredOrderHours = []int(hours)
	m.PreferredOrderWeekdays = []int(weekdays)
	m.BrandAffinity = map[string]any(brand)
	m.CategoryAffinity = map[string]any(category)
	m.UpdatedAt = m.UpdatedAt.UTC()
	m.CreatedAt = m.CreatedAt.UTC()
	return m, nil
}

func (r *AIModelRepo) Upsert(ctx context.Context, m domain.AICustomerModel) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ai_customer_models (
			profile_id, frequency, avg_order_value_minor, churn_prob, preferred_order_hours,
			preferred_order_weekdays, price_sensitivity, brand_affinity, category_affinity,
			model_version, updated_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (profile_id) DO UPDATE SET
			frequency=EXCLUDED.frequency, avg_order_value_minor=EXCLUDED.avg_order_value_minor,
			churn_prob=EXCLUDED.churn_prob, preferred_order_hours=EXCLUDED.preferred_order_hours,
			preferred_order_weekdays=EXCLUDED.preferred_order_weekdays,
			price_sensitivity=EXCLUDED.price_sensitivity, brand_affinity=EXCLUDED.brand_affinity,
			category_affinity=EXCLUDED.category_affinity, model_version=EXCLUDED.model_version,
			updated_at=EXCLUDED.updated_at`,
		m.ProfileID, m.Frequency, m.AvgOrderValueMinor, m.ChurnProb, IntArray(m.PreferredOrderHours),
		IntArray(m.PreferredOrderWeekdays), m.PriceSensitivity, JSONMap(metaGetMap(m.BrandAffinity)),
		JSONMap(metaGetMap(m.CategoryAffinity)), m.ModelVersion, m.UpdatedAt, m.CreatedAt,
	)
	return err
}
