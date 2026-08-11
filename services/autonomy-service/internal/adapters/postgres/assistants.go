package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type AssistantRepo struct{ DB *sql.DB }

var _ ports.AssistantRepo = (*AssistantRepo)(nil)

func (r *AssistantRepo) Save(ctx context.Context, a domain.ExecutiveAssistant) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_assistants (id, tenant_id, role, name, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, role=EXCLUDED.role, name=EXCLUDED.name,
			active=EXCLUDED.active, created_at=EXCLUDED.created_at`,
		a.ID, a.TenantID, string(a.Role), a.Name, a.Active, a.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *AssistantRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ExecutiveAssistant, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, role, name, active, created_at
		FROM auto_assistants WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ExecutiveAssistant, 0)
	for rows.Next() {
		var a domain.ExecutiveAssistant
		var role string
		if err := rows.Scan(&a.ID, &a.TenantID, &role, &a.Name, &a.Active, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Role = domain.AssistantRole(role)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}
