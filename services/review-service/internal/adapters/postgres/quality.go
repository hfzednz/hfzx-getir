package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// QualityRepo persists quality dimensions.
type QualityRepo struct{ DB *sql.DB }

func (r *QualityRepo) Save(ctx context.Context, q domain.QualityScore) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO quality_scores (id, review_id, tenant_id, dimension, value, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (review_id, dimension) DO UPDATE SET
		  id=EXCLUDED.id, value=EXCLUDED.value, created_at=EXCLUDED.created_at, tenant_id=EXCLUDED.tenant_id`,
		q.ID, q.ReviewID, q.TenantID, q.Dimension, q.Value, q.CreatedAt.UTC())
	return err
}

func (r *QualityRepo) ListByReview(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.QualityScore, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, review_id, tenant_id, dimension, value, created_at
		FROM quality_scores WHERE tenant_id=$1 AND review_id=$2`, tenantID, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.QualityScore, 0)
	for rows.Next() {
		var q domain.QualityScore
		if err := rows.Scan(&q.ID, &q.ReviewID, &q.TenantID, &q.Dimension, &q.Value, &q.CreatedAt); err != nil {
			return nil, err
		}
		q.CreatedAt = q.CreatedAt.UTC()
		out = append(out, q)
	}
	return out, rows.Err()
}

var _ ports.QualityRepo = (*QualityRepo)(nil)
