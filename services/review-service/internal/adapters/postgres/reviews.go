package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// ReviewRepo persists reviews and revisions.
type ReviewRepo struct{ DB *sql.DB }

func (r *ReviewRepo) Save(ctx context.Context, rev domain.Review) error {
	topics := TextArray(rev.Topics)
	tags := TextArray(rev.Tags)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO reviews (
		  id, tenant_id, author_id, target_type, target_id, order_id, locale, title, body,
		  anonymous, verified_purchase, verified_delivery, status, sentiment, topics, tags,
		  helpful_count, not_helpful_count, report_count, pinned, revision, idempotency_key,
		  created_at, updated_at, published_at, deleted_at
		) VALUES (
		  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26
		)
		ON CONFLICT (id) DO UPDATE SET
		  author_id=EXCLUDED.author_id, target_type=EXCLUDED.target_type, target_id=EXCLUDED.target_id,
		  order_id=EXCLUDED.order_id, locale=EXCLUDED.locale, title=EXCLUDED.title, body=EXCLUDED.body,
		  anonymous=EXCLUDED.anonymous, verified_purchase=EXCLUDED.verified_purchase,
		  verified_delivery=EXCLUDED.verified_delivery, status=EXCLUDED.status, sentiment=EXCLUDED.sentiment,
		  topics=EXCLUDED.topics, tags=EXCLUDED.tags, helpful_count=EXCLUDED.helpful_count,
		  not_helpful_count=EXCLUDED.not_helpful_count, report_count=EXCLUDED.report_count,
		  pinned=EXCLUDED.pinned, revision=EXCLUDED.revision, idempotency_key=EXCLUDED.idempotency_key,
		  updated_at=EXCLUDED.updated_at, published_at=EXCLUDED.published_at, deleted_at=EXCLUDED.deleted_at`,
		rev.ID, rev.TenantID, rev.AuthorID, rev.TargetType, rev.TargetID, nullUUID(rev.OrderID),
		rev.Locale, rev.Title, rev.Body, rev.Anonymous, rev.VerifiedPurchase, rev.VerifiedDelivery,
		rev.Status, rev.Sentiment, topics, tags, rev.HelpfulCount, rev.NotHelpfulCount, rev.ReportCount,
		rev.Pinned, rev.Revision, rev.IdempotencyKey, rev.CreatedAt.UTC(), rev.UpdatedAt.UTC(),
		nullTime(rev.PublishedAt), nullTime(rev.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *ReviewRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Review, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, author_id, target_type, target_id, order_id, locale, title, body,
		       anonymous, verified_purchase, verified_delivery, status, sentiment, topics, tags,
		       helpful_count, not_helpful_count, report_count, pinned, revision, idempotency_key,
		       created_at, updated_at, published_at, deleted_at
		FROM reviews WHERE id=$1 AND tenant_id=$2`, id, tenantID)
}

func (r *ReviewRepo) GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Review, bool, error) {
	rev, err := r.scanOne(ctx, `
		SELECT id, tenant_id, author_id, target_type, target_id, order_id, locale, title, body,
		       anonymous, verified_purchase, verified_delivery, status, sentiment, topics, tags,
		       helpful_count, not_helpful_count, report_count, pinned, revision, idempotency_key,
		       created_at, updated_at, published_at, deleted_at
		FROM reviews WHERE tenant_id=$1 AND idempotency_key=$2 AND idempotency_key <> ''`, tenantID, key)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.Review{}, false, nil
		}
		return domain.Review{}, false, err
	}
	return rev, true, nil
}

func (r *ReviewRepo) ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, status string, limit int) ([]domain.Review, error) {
	q := `
		SELECT id, tenant_id, author_id, target_type, target_id, order_id, locale, title, body,
		       anonymous, verified_purchase, verified_delivery, status, sentiment, topics, tags,
		       helpful_count, not_helpful_count, report_count, pinned, revision, idempotency_key,
		       created_at, updated_at, published_at, deleted_at
		FROM reviews
		WHERE tenant_id=$1 AND target_type=$2 AND target_id=$3`
	args := []any{tenantID, targetType, targetID}
	if status != "" {
		q += ` AND status=$4`
		args = append(args, status)
		if limit > 0 {
			q += ` LIMIT $5`
			args = append(args, limit)
		}
	} else if limit > 0 {
		q += ` LIMIT $4`
		args = append(args, limit)
	}
	return r.scanMany(ctx, q, args...)
}

func (r *ReviewRepo) ListByAuthor(ctx context.Context, tenantID, authorID uuid.UUID, limit int) ([]domain.Review, error) {
	q := `
		SELECT id, tenant_id, author_id, target_type, target_id, order_id, locale, title, body,
		       anonymous, verified_purchase, verified_delivery, status, sentiment, topics, tags,
		       helpful_count, not_helpful_count, report_count, pinned, revision, idempotency_key,
		       created_at, updated_at, published_at, deleted_at
		FROM reviews WHERE tenant_id=$1 AND author_id=$2 ORDER BY created_at DESC`
	args := []any{tenantID, authorID}
	if limit > 0 {
		q += ` LIMIT $3`
		args = append(args, limit)
	}
	return r.scanMany(ctx, q, args...)
}

func (r *ReviewRepo) SaveRevision(ctx context.Context, rev domain.ReviewRevision) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO review_revisions (id, review_id, tenant_id, revision, title, body, locale, created_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (review_id, revision) DO NOTHING`,
		rev.ID, rev.ReviewID, rev.TenantID, rev.Revision, rev.Title, rev.Body, rev.Locale,
		rev.CreatedAt.UTC(), rev.CreatedBy)
	return err
}

func (r *ReviewRepo) ListRevisions(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewRevision, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, review_id, tenant_id, revision, title, body, locale, created_at, created_by
		FROM review_revisions WHERE tenant_id=$1 AND review_id=$2 ORDER BY revision ASC`, tenantID, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReviewRevision, 0)
	for rows.Next() {
		var rev domain.ReviewRevision
		if err := rows.Scan(&rev.ID, &rev.ReviewID, &rev.TenantID, &rev.Revision, &rev.Title, &rev.Body,
			&rev.Locale, &rev.CreatedAt, &rev.CreatedBy); err != nil {
			return nil, err
		}
		rev.CreatedAt = rev.CreatedAt.UTC()
		out = append(out, rev)
	}
	return out, rows.Err()
}

func (r *ReviewRepo) CountRecentByAuthor(ctx context.Context, tenantID, authorID uuid.UUID, since time.Time) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reviews
		WHERE tenant_id=$1 AND author_id=$2 AND created_at >= $3`, tenantID, authorID, since.UTC()).Scan(&n)
	return n, err
}

func (r *ReviewRepo) CountDupBody(ctx context.Context, tenantID uuid.UUID, bodyHash string, since time.Time) (int, error) {
	var n int
	// Match memory hash: sha256(lower(trim(body))) as hex.
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reviews
		WHERE tenant_id=$1 AND created_at >= $2
		  AND encode(sha256(convert_to(lower(btrim(body)), 'UTF8')), 'hex') = $3`,
		tenantID, since.UTC(), bodyHash).Scan(&n)
	return n, err
}

func (r *ReviewRepo) scanOne(ctx context.Context, q string, args ...any) (domain.Review, error) {
	var rev domain.Review
	var orderID uuid.NullUUID
	var topics, tags TextArray
	var published, deleted sql.NullTime
	err := r.DB.QueryRowContext(ctx, q, args...).Scan(
		&rev.ID, &rev.TenantID, &rev.AuthorID, &rev.TargetType, &rev.TargetID, &orderID,
		&rev.Locale, &rev.Title, &rev.Body, &rev.Anonymous, &rev.VerifiedPurchase, &rev.VerifiedDelivery,
		&rev.Status, &rev.Sentiment, &topics, &tags, &rev.HelpfulCount, &rev.NotHelpfulCount, &rev.ReportCount,
		&rev.Pinned, &rev.Revision, &rev.IdempotencyKey, &rev.CreatedAt, &rev.UpdatedAt, &published, &deleted)
	if err != nil {
		return domain.Review{}, mapNotFound(err)
	}
	rev.OrderID = scanNullUUID(orderID)
	rev.Topics = []string(topics)
	rev.Tags = []string(tags)
	rev.CreatedAt = rev.CreatedAt.UTC()
	rev.UpdatedAt = rev.UpdatedAt.UTC()
	rev.PublishedAt = scanNullTime(published)
	rev.DeletedAt = scanNullTime(deleted)
	return rev, nil
}

func (r *ReviewRepo) scanMany(ctx context.Context, q string, args ...any) ([]domain.Review, error) {
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Review, 0)
	for rows.Next() {
		var rev domain.Review
		var orderID uuid.NullUUID
		var topics, tags TextArray
		var published, deleted sql.NullTime
		if err := rows.Scan(
			&rev.ID, &rev.TenantID, &rev.AuthorID, &rev.TargetType, &rev.TargetID, &orderID,
			&rev.Locale, &rev.Title, &rev.Body, &rev.Anonymous, &rev.VerifiedPurchase, &rev.VerifiedDelivery,
			&rev.Status, &rev.Sentiment, &topics, &tags, &rev.HelpfulCount, &rev.NotHelpfulCount, &rev.ReportCount,
			&rev.Pinned, &rev.Revision, &rev.IdempotencyKey, &rev.CreatedAt, &rev.UpdatedAt, &published, &deleted); err != nil {
			return nil, err
		}
		rev.OrderID = scanNullUUID(orderID)
		rev.Topics = []string(topics)
		rev.Tags = []string(tags)
		rev.CreatedAt = rev.CreatedAt.UTC()
		rev.UpdatedAt = rev.UpdatedAt.UTC()
		rev.PublishedAt = scanNullTime(published)
		rev.DeletedAt = scanNullTime(deleted)
		out = append(out, rev)
	}
	return out, rows.Err()
}

var _ ports.ReviewRepo = (*ReviewRepo)(nil)
