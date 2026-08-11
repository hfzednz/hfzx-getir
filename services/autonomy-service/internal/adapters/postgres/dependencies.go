package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type DependencyRepo struct{ DB *sql.DB }

var _ ports.DependencyRepo = (*DependencyRepo)(nil)

func (r *DependencyRepo) Save(ctx context.Context, e domain.DependencyEdge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_dependencies (id, tenant_id, from_service, to_service, relation, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, from_service=EXCLUDED.from_service,
			to_service=EXCLUDED.to_service, relation=EXCLUDED.relation, created_at=EXCLUDED.created_at`,
		e.ID, e.TenantID, e.FromService, e.ToService, e.Relation, e.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *DependencyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DependencyEdge, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, from_service, to_service, relation, created_at
		FROM auto_dependencies WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.DependencyEdge, 0)
	for rows.Next() {
		var e domain.DependencyEdge
		if err := rows.Scan(&e.ID, &e.TenantID, &e.FromService, &e.ToService, &e.Relation, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}
