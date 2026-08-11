package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type EvolutionRepo struct{ DB *sql.DB }

var _ ports.EvolutionRepo = (*EvolutionRepo)(nil)

func (r *EvolutionRepo) Save(ctx context.Context, t domain.EvolutionTask) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_evolution (id, tenant_id, kind, title, priority, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, kind=EXCLUDED.kind, title=EXCLUDED.title,
			priority=EXCLUDED.priority, status=EXCLUDED.status, created_at=EXCLUDED.created_at`,
		t.ID, t.TenantID, string(t.Kind), t.Title, t.Priority, t.Status, t.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *EvolutionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.EvolutionTask, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, title, priority, status, created_at
		FROM auto_evolution WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.EvolutionTask, 0)
	for rows.Next() {
		var t domain.EvolutionTask
		var kind string
		if err := rows.Scan(&t.ID, &t.TenantID, &kind, &t.Title, &t.Priority, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Kind = domain.EvolutionKind(kind)
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}
