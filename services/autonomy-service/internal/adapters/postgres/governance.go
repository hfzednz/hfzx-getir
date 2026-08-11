package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type GovernanceRepo struct{ DB *sql.DB }

var _ ports.GovernanceRepo = (*GovernanceRepo)(nil)

func (r *GovernanceRepo) Save(ctx context.Context, g domain.GovernanceLoop) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_governance (id, tenant_id, domain, cadence, healthy, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, domain=EXCLUDED.domain, cadence=EXCLUDED.cadence,
			healthy=EXCLUDED.healthy, updated_at=EXCLUDED.updated_at`,
		g.ID, g.TenantID, g.Domain, g.Cadence, g.Healthy, g.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *GovernanceRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GovernanceLoop, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, domain, cadence, healthy, updated_at
		FROM auto_governance WHERE tenant_id=$1 ORDER BY updated_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.GovernanceLoop, 0)
	for rows.Next() {
		var g domain.GovernanceLoop
		if err := rows.Scan(&g.ID, &g.TenantID, &g.Domain, &g.Cadence, &g.Healthy, &g.UpdatedAt); err != nil {
			return nil, err
		}
		g.UpdatedAt = g.UpdatedAt.UTC()
		out = append(out, g)
	}
	return out, rows.Err()
}
