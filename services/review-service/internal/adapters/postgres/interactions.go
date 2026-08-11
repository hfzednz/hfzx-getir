package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// InteractionRepo votes, comments, reports.
type InteractionRepo struct{ DB *sql.DB }

func (r *InteractionRepo) SaveVote(ctx context.Context, v domain.ReviewVote) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO review_votes (id, review_id, tenant_id, voter_id, helpful, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (review_id, voter_id) DO UPDATE SET
		  id=EXCLUDED.id, helpful=EXCLUDED.helpful, created_at=EXCLUDED.created_at, tenant_id=EXCLUDED.tenant_id`,
		v.ID, v.ReviewID, v.TenantID, v.VoterID, v.Helpful, v.CreatedAt.UTC())
	return err
}

func (r *InteractionRepo) GetVote(ctx context.Context, reviewID, voterID uuid.UUID) (domain.ReviewVote, bool, error) {
	var v domain.ReviewVote
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, review_id, tenant_id, voter_id, helpful, created_at
		FROM review_votes WHERE review_id=$1 AND voter_id=$2`, reviewID, voterID).Scan(
		&v.ID, &v.ReviewID, &v.TenantID, &v.VoterID, &v.Helpful, &v.CreatedAt)
	if isNoRows(err) {
		return domain.ReviewVote{}, false, nil
	}
	if err != nil {
		return domain.ReviewVote{}, false, err
	}
	v.CreatedAt = v.CreatedAt.UTC()
	return v, true, nil
}

func (r *InteractionRepo) SaveComment(ctx context.Context, c domain.ReviewComment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO review_comments
		  (id, review_id, tenant_id, author_id, parent_id, body, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
		  body=EXCLUDED.body, status=EXCLUDED.status, updated_at=EXCLUDED.updated_at, parent_id=EXCLUDED.parent_id`,
		c.ID, c.ReviewID, c.TenantID, c.AuthorID, nullUUID(c.ParentID), c.Body, c.Status,
		c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return err
}

func (r *InteractionRepo) ListComments(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewComment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, review_id, tenant_id, author_id, parent_id, body, status, created_at, updated_at
		FROM review_comments WHERE tenant_id=$1 AND review_id=$2 ORDER BY created_at ASC`, tenantID, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReviewComment, 0)
	for rows.Next() {
		var c domain.ReviewComment
		var parent uuid.NullUUID
		if err := rows.Scan(&c.ID, &c.ReviewID, &c.TenantID, &c.AuthorID, &parent, &c.Body, &c.Status,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.ParentID = scanNullUUID(parent)
		c.CreatedAt = c.CreatedAt.UTC()
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *InteractionRepo) SaveReport(ctx context.Context, rep domain.ReviewReport) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO review_reports (id, review_id, tenant_id, reporter_id, reason, details, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rep.ID, rep.ReviewID, rep.TenantID, rep.ReporterID, rep.Reason, rep.Details, rep.CreatedAt.UTC())
	return err
}

var _ ports.InteractionRepo = (*InteractionRepo)(nil)
