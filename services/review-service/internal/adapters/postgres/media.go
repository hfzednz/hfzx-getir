package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// MediaRepo persists media refs.
type MediaRepo struct{ DB *sql.DB }

func (r *MediaRepo) Save(ctx context.Context, m domain.ReviewMedia) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO review_media
		  (id, review_id, tenant_id, media_ref, kind, mime_type, width, height, duration_ms,
		   verified, moderation_ok, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
		  media_ref=EXCLUDED.media_ref, kind=EXCLUDED.kind, mime_type=EXCLUDED.mime_type,
		  width=EXCLUDED.width, height=EXCLUDED.height, duration_ms=EXCLUDED.duration_ms,
		  verified=EXCLUDED.verified, moderation_ok=EXCLUDED.moderation_ok`,
		m.ID, m.ReviewID, m.TenantID, m.MediaRef, m.Kind, m.MimeType, m.Width, m.Height,
		m.DurationMs, m.Verified, m.ModerationOK, m.CreatedAt.UTC())
	return err
}

func (r *MediaRepo) ListByReview(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewMedia, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, review_id, tenant_id, media_ref, kind, mime_type, width, height, duration_ms,
		       verified, moderation_ok, created_at
		FROM review_media WHERE tenant_id=$1 AND review_id=$2 ORDER BY created_at ASC`, tenantID, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReviewMedia, 0)
	for rows.Next() {
		var m domain.ReviewMedia
		if err := rows.Scan(&m.ID, &m.ReviewID, &m.TenantID, &m.MediaRef, &m.Kind, &m.MimeType,
			&m.Width, &m.Height, &m.DurationMs, &m.Verified, &m.ModerationOK, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.MediaRepo = (*MediaRepo)(nil)
