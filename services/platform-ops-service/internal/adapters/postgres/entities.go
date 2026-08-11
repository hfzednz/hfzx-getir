package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/app/ports"
	"github.com/nexora/platform-ops-service/internal/domain"
)

// DeploymentRepo persists deployments.
type DeploymentRepo struct{ DB *sql.DB }

func (r *DeploymentRepo) Save(ctx context.Context, d domain.Deployment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_deployments (
			id, tenant_id, service, environment, strategy, image_tag, status, started_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			service=EXCLUDED.service, environment=EXCLUDED.environment, strategy=EXCLUDED.strategy,
			image_tag=EXCLUDED.image_tag, status=EXCLUDED.status, started_at=EXCLUDED.started_at,
			completed_at=EXCLUDED.completed_at`,
		d.ID, d.TenantID, d.Service, d.Environment, d.Strategy, d.ImageTag, d.Status,
		d.StartedAt.UTC(), nullTime(d.CompletedAt))
	return mapUniqueViolation(err)
}

func (r *DeploymentRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Deployment, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, service, environment, strategy, image_tag, status, started_at, completed_at
		FROM platform_deployments WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	var d domain.Deployment
	var completed sql.NullTime
	err := row.Scan(&d.ID, &d.TenantID, &d.Service, &d.Environment, &d.Strategy, &d.ImageTag, &d.Status,
		&d.StartedAt, &completed)
	if err != nil {
		if isNoRows(err) {
			return domain.Deployment{}, domain.ErrNotFound
		}
		return domain.Deployment{}, err
	}
	d.StartedAt = d.StartedAt.UTC()
	d.CompletedAt = scanNullTime(completed)
	return d, nil
}

func (r *DeploymentRepo) List(ctx context.Context, tenantID uuid.UUID, env string) ([]domain.Deployment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, service, environment, strategy, image_tag, status, started_at, completed_at
		FROM platform_deployments WHERE tenant_id=$1 AND ($2='' OR environment=$2)
		ORDER BY started_at DESC`, tenantID, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Deployment{}
	for rows.Next() {
		var d domain.Deployment
		var completed sql.NullTime
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Service, &d.Environment, &d.Strategy, &d.ImageTag, &d.Status,
			&d.StartedAt, &completed); err != nil {
			return nil, err
		}
		d.StartedAt = d.StartedAt.UTC()
		d.CompletedAt = scanNullTime(completed)
		out = append(out, d)
	}
	return out, rows.Err()
}

// ScalingRepo persists scaling events.
type ScalingRepo struct{ DB *sql.DB }

func (r *ScalingRepo) Save(ctx context.Context, s domain.ScalingEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_scaling_events (
			id, tenant_id, service, environment, from_replicas, to_replicas, reason, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		s.ID, s.TenantID, s.Service, s.Environment, s.FromReplicas, s.ToReplicas, s.Reason, s.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ScalingRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ScalingEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, service, environment, from_replicas, to_replicas, reason, created_at
		FROM platform_scaling_events WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ScalingEvent{}
	for rows.Next() {
		var s domain.ScalingEvent
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Service, &s.Environment, &s.FromReplicas, &s.ToReplicas,
			&s.Reason, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

// BackupRepo persists backup jobs.
type BackupRepo struct{ DB *sql.DB }

func (r *BackupRepo) Save(ctx context.Context, b domain.BackupJob) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_backups (
			id, tenant_id, kind, target, status, location, started_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, target=EXCLUDED.target, status=EXCLUDED.status,
			location=EXCLUDED.location, started_at=EXCLUDED.started_at, completed_at=EXCLUDED.completed_at`,
		b.ID, b.TenantID, b.Kind, b.Target, b.Status, b.Location, b.StartedAt.UTC(), nullTime(b.CompletedAt))
	return mapUniqueViolation(err)
}

func (r *BackupRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.BackupJob, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, target, status, location, started_at, completed_at
		FROM platform_backups WHERE tenant_id=$1 ORDER BY started_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BackupJob{}
	for rows.Next() {
		var b domain.BackupJob
		var completed sql.NullTime
		if err := rows.Scan(&b.ID, &b.TenantID, &b.Kind, &b.Target, &b.Status, &b.Location, &b.StartedAt, &completed); err != nil {
			return nil, err
		}
		b.StartedAt = b.StartedAt.UTC()
		b.CompletedAt = scanNullTime(completed)
		out = append(out, b)
	}
	return out, rows.Err()
}

// RecoveryRepo persists recovery jobs.
type RecoveryRepo struct{ DB *sql.DB }

func (r *RecoveryRepo) Save(ctx context.Context, rec domain.RecoveryJob) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_recoveries (
			id, tenant_id, kind, status, notes, started_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, status=EXCLUDED.status, notes=EXCLUDED.notes,
			started_at=EXCLUDED.started_at, completed_at=EXCLUDED.completed_at`,
		rec.ID, rec.TenantID, rec.Kind, rec.Status, rec.Notes, rec.StartedAt.UTC(), nullTime(rec.CompletedAt))
	return mapUniqueViolation(err)
}

func (r *RecoveryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RecoveryJob, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, status, notes, started_at, completed_at
		FROM platform_recoveries WHERE tenant_id=$1 ORDER BY started_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RecoveryJob{}
	for rows.Next() {
		var rec domain.RecoveryJob
		var completed sql.NullTime
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.Kind, &rec.Status, &rec.Notes, &rec.StartedAt, &completed); err != nil {
			return nil, err
		}
		rec.StartedAt = rec.StartedAt.UTC()
		rec.CompletedAt = scanNullTime(completed)
		out = append(out, rec)
	}
	return out, rows.Err()
}

// AlertRepo persists alert events.
type AlertRepo struct{ DB *sql.DB }

func (r *AlertRepo) Save(ctx context.Context, a domain.AlertEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_alerts (
			id, tenant_id, name, severity, status, labels_json, fired_at, resolved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, severity=EXCLUDED.severity, status=EXCLUDED.status,
			labels_json=EXCLUDED.labels_json, fired_at=EXCLUDED.fired_at, resolved_at=EXCLUDED.resolved_at`,
		a.ID, a.TenantID, a.Name, a.Severity, a.Status, JSONStringMap(a.Labels), a.FiredAt.UTC(), nullTime(a.ResolvedAt))
	return mapUniqueViolation(err)
}

func (r *AlertRepo) List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.AlertEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, severity, status, labels_json, fired_at, resolved_at
		FROM platform_alerts WHERE tenant_id=$1 AND ($2='' OR status=$2)
		ORDER BY fired_at DESC`, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AlertEvent{}
	for rows.Next() {
		var a domain.AlertEvent
		var labels JSONStringMap
		var resolved sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &a.Severity, &a.Status, &labels, &a.FiredAt, &resolved); err != nil {
			return nil, err
		}
		a.Labels = map[string]string(labels)
		a.FiredAt = a.FiredAt.UTC()
		a.ResolvedAt = scanNullTime(resolved)
		out = append(out, a)
	}
	return out, rows.Err()
}

// CostRepo persists cost snapshots.
type CostRepo struct{ DB *sql.DB }

func (r *CostRepo) Save(ctx context.Context, c domain.CostSnapshot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_costs (
			id, tenant_id, environment, amount_minor, currency, period, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.Environment, c.AmountMinor, c.Currency, c.Period, c.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *CostRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CostSnapshot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, environment, amount_minor, currency, period, created_at
		FROM platform_costs WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CostSnapshot{}
	for rows.Next() {
		var c domain.CostSnapshot
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Environment, &c.AmountMinor, &c.Currency, &c.Period, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

// SLORepo persists SLO reports.
type SLORepo struct{ DB *sql.DB }

func (r *SLORepo) Save(ctx context.Context, s domain.SLOReport) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_slo_reports (
			id, tenant_id, service, objective, actual, budget_left, window, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		s.ID, s.TenantID, s.Service, s.Objective, s.Actual, s.BudgetLeft, s.Window, s.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *SLORepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.SLOReport, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, service, objective, actual, budget_left, window, created_at
		FROM platform_slo_reports WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SLOReport{}
	for rows.Next() {
		var s domain.SLOReport
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Service, &s.Objective, &s.Actual, &s.BudgetLeft, &s.Window, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

var (
	_ ports.DeploymentRepo = (*DeploymentRepo)(nil)
	_ ports.ScalingRepo    = (*ScalingRepo)(nil)
	_ ports.BackupRepo     = (*BackupRepo)(nil)
	_ ports.RecoveryRepo   = (*RecoveryRepo)(nil)
	_ ports.AlertRepo      = (*AlertRepo)(nil)
	_ ports.CostRepo       = (*CostRepo)(nil)
	_ ports.SLORepo        = (*SLORepo)(nil)
)
