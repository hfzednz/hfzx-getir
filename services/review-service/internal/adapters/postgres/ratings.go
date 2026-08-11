package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// RatingRepo persists ratings and aggregates.
type RatingRepo struct{ DB *sql.DB }

func (r *RatingRepo) Save(ctx context.Context, rt domain.Rating) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ratings
		  (id, tenant_id, author_id, target_type, target_id, review_id, scheme, value, stars, verified, weight, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		rt.ID, rt.TenantID, rt.AuthorID, rt.TargetType, rt.TargetID, nullUUID(rt.ReviewID),
		rt.Scheme, rt.Value, rt.Stars, rt.Verified, rt.Weight, rt.CreatedAt.UTC())
	return err
}

func (r *RatingRepo) ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) ([]domain.Rating, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, author_id, target_type, target_id, review_id, scheme, value, stars, verified, weight, created_at
		FROM ratings WHERE tenant_id=$1 AND target_type=$2 AND target_id=$3
		ORDER BY created_at DESC`, tenantID, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Rating, 0)
	for rows.Next() {
		var rt domain.Rating
		var reviewID uuid.NullUUID
		if err := rows.Scan(&rt.ID, &rt.TenantID, &rt.AuthorID, &rt.TargetType, &rt.TargetID, &reviewID,
			&rt.Scheme, &rt.Value, &rt.Stars, &rt.Verified, &rt.Weight, &rt.CreatedAt); err != nil {
			return nil, err
		}
		rt.ReviewID = scanNullUUID(reviewID)
		rt.CreatedAt = rt.CreatedAt.UTC()
		out = append(out, rt)
	}
	return out, rows.Err()
}

func (r *RatingRepo) GetAggregate(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, scheme string) (domain.RatingAggregate, error) {
	var a domain.RatingAggregate
	err := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, target_type, target_id, scheme, count, sum_stars, avg_stars, bayesian_avg,
		       time_decay_avg, verified_count, verified_avg, updated_at
		FROM rating_aggregates
		WHERE tenant_id=$1 AND target_type=$2 AND target_id=$3 AND scheme=$4`,
		tenantID, targetType, targetID, scheme).Scan(
		&a.TenantID, &a.TargetType, &a.TargetID, &a.Scheme, &a.Count, &a.SumStars, &a.AvgStars,
		&a.BayesianAvg, &a.TimeDecayAvg, &a.VerifiedCount, &a.VerifiedAvg, &a.UpdatedAt)
	if err != nil {
		return domain.RatingAggregate{}, mapNotFound(err)
	}
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

func (r *RatingRepo) SaveAggregate(ctx context.Context, a domain.RatingAggregate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO rating_aggregates
		  (tenant_id, target_type, target_id, scheme, count, sum_stars, avg_stars, bayesian_avg,
		   time_decay_avg, verified_count, verified_avg, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, target_type, target_id, scheme) DO UPDATE SET
		  count=EXCLUDED.count, sum_stars=EXCLUDED.sum_stars, avg_stars=EXCLUDED.avg_stars,
		  bayesian_avg=EXCLUDED.bayesian_avg, time_decay_avg=EXCLUDED.time_decay_avg,
		  verified_count=EXCLUDED.verified_count, verified_avg=EXCLUDED.verified_avg,
		  updated_at=EXCLUDED.updated_at`,
		a.TenantID, a.TargetType, a.TargetID, a.Scheme, a.Count, a.SumStars, a.AvgStars,
		a.BayesianAvg, a.TimeDecayAvg, a.VerifiedCount, a.VerifiedAvg, a.UpdatedAt.UTC())
	return err
}

var _ ports.RatingRepo = (*RatingRepo)(nil)
