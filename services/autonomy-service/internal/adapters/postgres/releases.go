package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type ReleaseRepo struct{ DB *sql.DB }

var _ ports.ReleaseRepo = (*ReleaseRepo)(nil)

func (r *ReleaseRepo) Save(ctx context.Context, p domain.ReleasePlan) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_releases (id, tenant_id, version, strategy, score, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, version=EXCLUDED.version, strategy=EXCLUDED.strategy,
			score=EXCLUDED.score, status=EXCLUDED.status, created_at=EXCLUDED.created_at`,
		p.ID, p.TenantID, p.Version, p.Strategy, p.Score, p.Status, p.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *ReleaseRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ReleasePlan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, version, strategy, score, status, created_at
		FROM auto_releases WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReleasePlan, 0)
	for rows.Next() {
		var p domain.ReleasePlan
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Version, &p.Strategy, &p.Score, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}
