package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type AuditRepo struct{ DB *sql.DB }

var _ ports.AuditRepo = (*AuditRepo)(nil)

func (r *AuditRepo) Save(ctx context.Context, a domain.AutonomyAudit) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_audits (id, tenant_id, scope, status, score, created_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, scope=EXCLUDED.scope, status=EXCLUDED.status,
			score=EXCLUDED.score, created_at=EXCLUDED.created_at, completed_at=EXCLUDED.completed_at`,
		a.ID, a.TenantID, string(a.Scope), a.Status, a.Score, a.CreatedAt, nullTime(a.CompletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *AuditRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.AutonomyAudit, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, scope, status, score, created_at, completed_at
		FROM auto_audits WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.AutonomyAudit, 0)
	for rows.Next() {
		var a domain.AutonomyAudit
		var scope string
		var completed sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &scope, &a.Status, &a.Score, &a.CreatedAt, &completed); err != nil {
			return nil, err
		}
		a.Scope = domain.AuditScope(scope)
		a.CompletedAt = scanNullTime(completed)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type WeaknessRepo struct{ DB *sql.DB }

var _ ports.WeaknessRepo = (*WeaknessRepo)(nil)

func (r *WeaknessRepo) Save(ctx context.Context, w domain.Weakness) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_weaknesses (
			id, tenant_id, audit_id, code, title, severity, status, resolution, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id=EXCLUDED.tenant_id, audit_id=EXCLUDED.audit_id, code=EXCLUDED.code,
			title=EXCLUDED.title, severity=EXCLUDED.severity, status=EXCLUDED.status,
			resolution=EXCLUDED.resolution, created_at=EXCLUDED.created_at`,
		w.ID, w.TenantID, w.AuditID, w.Code, w.Title, w.Severity, w.Status, w.Resolution, w.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *WeaknessRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Weakness, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, audit_id, code, title, severity, status, resolution, created_at
		FROM auto_weaknesses WHERE tenant_id=$1 ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Weakness, 0)
	for rows.Next() {
		var w domain.Weakness
		if err := rows.Scan(
			&w.ID, &w.TenantID, &w.AuditID, &w.Code, &w.Title, &w.Severity, &w.Status, &w.Resolution, &w.CreatedAt,
		); err != nil {
			return nil, err
		}
		w.CreatedAt = w.CreatedAt.UTC()
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *WeaknessRepo) OpenCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM auto_weaknesses WHERE tenant_id=$1 AND status='open'`, tenantID).Scan(&n)
	return n, err
}
