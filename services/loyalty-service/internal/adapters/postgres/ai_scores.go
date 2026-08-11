package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// AIScoreRepo persists stub AI scores.
type AIScoreRepo struct{ DB *sql.DB }

func (r *AIScoreRepo) Upsert(ctx context.Context, s domain.AIScore) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO ai_scores (id, tenant_id, account_id, principal_id, churn_score, ltv_score, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, account_id) DO UPDATE SET
		  principal_id = EXCLUDED.principal_id,
		  churn_score = EXCLUDED.churn_score,
		  ltv_score = EXCLUDED.ltv_score,
		  updated_at = EXCLUDED.updated_at`,
		s.ID, s.TenantID, s.AccountID, s.PrincipalID, s.ChurnScore, s.LTVScore, s.UpdatedAt.UTC())
	return err
}

func (r *AIScoreRepo) Get(ctx context.Context, tenantID, accountID uuid.UUID) (domain.AIScore, error) {
	var s domain.AIScore
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, principal_id, churn_score, ltv_score, updated_at
		FROM ai_scores WHERE tenant_id=$1 AND account_id=$2`, tenantID, accountID).Scan(
		&s.ID, &s.TenantID, &s.AccountID, &s.PrincipalID, &s.ChurnScore, &s.LTVScore, &s.UpdatedAt)
	if isNoRows(err) {
		return domain.AIScore{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.AIScore{}, err
	}
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

var _ ports.AIScoreRepo = (*AIScoreRepo)(nil)
