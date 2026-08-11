package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app/ports"
	"github.com/nexora/recommendation-service/internal/domain"
)

// FeatureRepo persists product feature vectors.
type FeatureRepo struct{ DB *sql.DB }

func (r *FeatureRepo) Upsert(ctx context.Context, f domain.ProductFeatures) error {
	tags := TextArray(f.Tags)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_features
		  (product_id, category_id, brand_id, tags, price_minor, popularity, rating_avg)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (product_id) DO UPDATE SET
		  category_id=EXCLUDED.category_id,
		  brand_id=EXCLUDED.brand_id,
		  tags=EXCLUDED.tags,
		  price_minor=EXCLUDED.price_minor,
		  popularity=EXCLUDED.popularity,
		  rating_avg=EXCLUDED.rating_avg`,
		f.ProductID, nullUUIDValue(f.CategoryID), nullUUIDValue(f.BrandID), tags,
		f.PriceMinor, f.Popularity, f.RatingAvg)
	return err
}

func (r *FeatureRepo) Get(ctx context.Context, productID uuid.UUID) (domain.ProductFeatures, error) {
	var f domain.ProductFeatures
	var category, brand uuid.NullUUID
	var tags TextArray
	err := r.DB.QueryRowContext(ctx, `
		SELECT product_id, category_id, brand_id, tags, price_minor, popularity, rating_avg
		FROM product_features WHERE product_id=$1`, productID).Scan(
		&f.ProductID, &category, &brand, &tags, &f.PriceMinor, &f.Popularity, &f.RatingAvg)
	if err != nil {
		return domain.ProductFeatures{}, mapNotFound(err)
	}
	f.CategoryID = scanUUIDOrNil(category)
	f.BrandID = scanUUIDOrNil(brand)
	f.Tags = []string(tags)
	return f, nil
}

func (r *FeatureRepo) List(ctx context.Context, ids []uuid.UUID) ([]domain.ProductFeatures, error) {
	out := make([]domain.ProductFeatures, 0, len(ids))
	for _, id := range ids {
		f, err := r.Get(ctx, id)
		if err != nil {
			if isNoRows(err) || err == domain.ErrNotFound {
				continue
			}
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *FeatureRepo) ListAll(ctx context.Context, limit int) ([]domain.ProductFeatures, error) {
	q := `
		SELECT product_id, category_id, brand_id, tags, price_minor, popularity, rating_avg
		FROM product_features`
	args := []any{}
	if limit > 0 {
		q += ` LIMIT $1`
		args = append(args, limit)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ProductFeatures, 0)
	for rows.Next() {
		var f domain.ProductFeatures
		var category, brand uuid.NullUUID
		var tags TextArray
		if err := rows.Scan(&f.ProductID, &category, &brand, &tags, &f.PriceMinor, &f.Popularity, &f.RatingAvg); err != nil {
			return nil, err
		}
		f.CategoryID = scanUUIDOrNil(category)
		f.BrandID = scanUUIDOrNil(brand)
		f.Tags = []string(tags)
		out = append(out, f)
	}
	return out, rows.Err()
}

var _ ports.FeatureRepo = (*FeatureRepo)(nil)
