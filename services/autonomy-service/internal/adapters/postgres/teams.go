package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type TeamRepo struct{ DB *sql.DB }

var _ ports.TeamRepo = (*TeamRepo)(nil)

func (r *TeamRepo) Save(ctx context.Context, t domain.DigitalTeam) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_teams (id, tenant_id, kind, name, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, kind=EXCLUDED.kind, name=EXCLUDED.name,
			active=EXCLUDED.active, created_at=EXCLUDED.created_at`,
		t.ID, t.TenantID, string(t.Kind), t.Name, t.Active, t.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *TeamRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DigitalTeam, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, name, active, created_at
		FROM auto_teams WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.DigitalTeam, 0)
	for rows.Next() {
		var t domain.DigitalTeam
		var kind string
		if err := rows.Scan(&t.ID, &t.TenantID, &kind, &t.Name, &t.Active, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Kind = domain.DigitalTeamKind(kind)
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}
