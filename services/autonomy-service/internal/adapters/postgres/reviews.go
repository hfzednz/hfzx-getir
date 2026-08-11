package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type ReviewRepo struct{ DB *sql.DB }

var _ ports.ReviewRepo = (*ReviewRepo)(nil)

func (r *ReviewRepo) Save(ctx context.Context, rev domain.AICTOReview) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_reviews (id, tenant_id, kind, summary, debt_score, suggestions, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, kind=EXCLUDED.kind, summary=EXCLUDED.summary,
			debt_score=EXCLUDED.debt_score, suggestions=EXCLUDED.suggestions, created_at=EXCLUDED.created_at`,
		rev.ID, rev.TenantID, string(rev.Kind), rev.Summary, rev.DebtScore,
		JSONStrings(rev.Suggestions), rev.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *ReviewRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.AICTOReview, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, summary, debt_score, suggestions, created_at
		FROM auto_reviews WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.AICTOReview, 0)
	for rows.Next() {
		var rev domain.AICTOReview
		var kind string
		var suggestions JSONStrings
		if err := rows.Scan(
			&rev.ID, &rev.TenantID, &kind, &rev.Summary, &rev.DebtScore, &suggestions, &rev.CreatedAt,
		); err != nil {
			return nil, err
		}
		rev.Kind = domain.ReviewKind(kind)
		rev.Suggestions = []string(suggestions)
		if rev.Suggestions == nil {
			rev.Suggestions = []string{}
		}
		rev.CreatedAt = rev.CreatedAt.UTC()
		out = append(out, rev)
	}
	return out, rows.Err()
}
