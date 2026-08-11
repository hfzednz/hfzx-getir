package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type HealRepo struct{ DB *sql.DB }

var _ ports.HealRepo = (*HealRepo)(nil)

func (r *HealRepo) Save(ctx context.Context, a domain.HealAction) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_heals (
			id, tenant_id, kind, target_ref, action, status, automated, created_at, executed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, kind=EXCLUDED.kind, target_ref=EXCLUDED.target_ref,
			action=EXCLUDED.action, status=EXCLUDED.status, automated=EXCLUDED.automated,
			created_at=EXCLUDED.created_at, executed_at=EXCLUDED.executed_at`,
		a.ID, a.TenantID, string(a.Kind), a.TargetRef, a.Action, a.Status, a.Automated,
		a.CreatedAt, nullTime(a.ExecutedAt),
	)
	return mapUniqueViolation(err)
}

func (r *HealRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.HealAction, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, target_ref, action, status, automated, created_at, executed_at
		FROM auto_heals WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.HealAction, 0)
	for rows.Next() {
		var a domain.HealAction
		var kind string
		var executed sql.NullTime
		if err := rows.Scan(
			&a.ID, &a.TenantID, &kind, &a.TargetRef, &a.Action, &a.Status, &a.Automated,
			&a.CreatedAt, &executed,
		); err != nil {
			return nil, err
		}
		a.Kind = domain.HealKind(kind)
		a.ExecutedAt = scanNullTime(executed)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *HealRepo) ExecutedCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM auto_heals WHERE tenant_id=$1 AND status='executed'`, tenantID).Scan(&n)
	return n, err
}
