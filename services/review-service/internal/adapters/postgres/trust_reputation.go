package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
)

// TrustRepo persists reviewer trust.
type TrustRepo struct{ DB *sql.DB }

func (r *TrustRepo) Get(ctx context.Context, tenantID, reviewerID uuid.UUID) (domain.TrustScore, error) {
	var t domain.TrustScore
	var badges TextArray
	err := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, reviewer_id, score, verified_purchases, published_reviews, rejected_reviews,
		       helpful_received, badges, updated_at
		FROM trust_scores WHERE tenant_id=$1 AND reviewer_id=$2`, tenantID, reviewerID).Scan(
		&t.TenantID, &t.ReviewerID, &t.Score, &t.VerifiedPurchases, &t.PublishedReviews, &t.RejectedReviews,
		&t.HelpfulReceived, &badges, &t.UpdatedAt)
	if err != nil {
		return domain.TrustScore{}, mapNotFound(err)
	}
	t.Badges = []string(badges)
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

func (r *TrustRepo) Save(ctx context.Context, t domain.TrustScore) error {
	badges := TextArray(t.Badges)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO trust_scores
		  (tenant_id, reviewer_id, score, verified_purchases, published_reviews, rejected_reviews,
		   helpful_received, badges, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, reviewer_id) DO UPDATE SET
		  score=EXCLUDED.score, verified_purchases=EXCLUDED.verified_purchases,
		  published_reviews=EXCLUDED.published_reviews, rejected_reviews=EXCLUDED.rejected_reviews,
		  helpful_received=EXCLUDED.helpful_received, badges=EXCLUDED.badges, updated_at=EXCLUDED.updated_at`,
		t.TenantID, t.ReviewerID, t.Score, t.VerifiedPurchases, t.PublishedReviews, t.RejectedReviews,
		t.HelpfulReceived, badges, t.UpdatedAt.UTC())
	return err
}

var _ ports.TrustRepo = (*TrustRepo)(nil)

// ReputationRepo persists entity reputation.
type ReputationRepo struct{ DB *sql.DB }

func (r *ReputationRepo) Get(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) (domain.ReputationScore, error) {
	var rep domain.ReputationScore
	err := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, target_type, target_id, score, tier, review_count, updated_at
		FROM reputation_scores WHERE tenant_id=$1 AND target_type=$2 AND target_id=$3`,
		tenantID, targetType, targetID).Scan(
		&rep.TenantID, &rep.TargetType, &rep.TargetID, &rep.Score, &rep.Tier, &rep.ReviewCount, &rep.UpdatedAt)
	if err != nil {
		return domain.ReputationScore{}, mapNotFound(err)
	}
	rep.UpdatedAt = rep.UpdatedAt.UTC()
	return rep, nil
}

func (r *ReputationRepo) Save(ctx context.Context, rep domain.ReputationScore) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO reputation_scores
		  (tenant_id, target_type, target_id, score, tier, review_count, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, target_type, target_id) DO UPDATE SET
		  score=EXCLUDED.score, tier=EXCLUDED.tier, review_count=EXCLUDED.review_count,
		  updated_at=EXCLUDED.updated_at`,
		rep.TenantID, rep.TargetType, rep.TargetID, rep.Score, rep.Tier, rep.ReviewCount, rep.UpdatedAt.UTC())
	return err
}

var _ ports.ReputationRepo = (*ReputationRepo)(nil)
