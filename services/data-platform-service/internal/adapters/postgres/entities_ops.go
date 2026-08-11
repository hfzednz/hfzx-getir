package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/data-platform-service/internal/app/ports"
	"github.com/nexora/data-platform-service/internal/domain"
)

type ExperimentRepo struct{ DB *sql.DB }

func (r *ExperimentRepo) Save(ctx context.Context, e domain.Experiment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO experiments (
			id, tenant_id, key, name, status, variants, primary_kpi, started_at, ended_at, winner, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, name=EXCLUDED.name, status=EXCLUDED.status, variants=EXCLUDED.variants,
			primary_kpi=EXCLUDED.primary_kpi, started_at=EXCLUDED.started_at, ended_at=EXCLUDED.ended_at,
			winner=EXCLUDED.winner, updated_at=EXCLUDED.updated_at`,
		e.ID, e.TenantID, e.Key, e.Name, e.Status, VariantsJSON(e.Variants), e.PrimaryKPI,
		nullTime(e.StartedAt), nullTime(e.EndedAt), e.Winner, e.UpdatedAt.UTC())
	return err
}

func (r *ExperimentRepo) Get(ctx context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, status, variants, primary_kpi, started_at, ended_at, winner, updated_at
		FROM experiments WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var e domain.Experiment
	var variants VariantsJSON
	var started, ended sql.NullTime
	err := row.Scan(&e.ID, &e.TenantID, &e.Key, &e.Name, &e.Status, &variants, &e.PrimaryKPI,
		&started, &ended, &e.Winner, &e.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Experiment{}, domain.ErrNotFound
		}
		return domain.Experiment{}, err
	}
	e.Variants = []domain.ExperimentVariant(variants)
	e.StartedAt = scanNullTime(started)
	e.EndedAt = scanNullTime(ended)
	e.UpdatedAt = e.UpdatedAt.UTC()
	return e, nil
}

func (r *ExperimentRepo) SaveAssignment(ctx context.Context, a domain.ExperimentAssignment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO experiment_assignments (tenant_id, experiment_id, subject_id, variant, assigned_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_id, experiment_id, subject_id) DO UPDATE SET
			variant=EXCLUDED.variant, assigned_at=EXCLUDED.assigned_at`,
		a.TenantID, a.ExperimentID, a.SubjectID, a.Variant, a.AssignedAt.UTC())
	return err
}

func (r *ExperimentRepo) GetAssignment(ctx context.Context, tenantID, experimentID, subjectID uuid.UUID) (domain.ExperimentAssignment, bool, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, experiment_id, subject_id, variant, assigned_at
		FROM experiment_assignments WHERE tenant_id=$1 AND experiment_id=$2 AND subject_id=$3`,
		tenantID, experimentID, subjectID)
	var a domain.ExperimentAssignment
	err := row.Scan(&a.TenantID, &a.ExperimentID, &a.SubjectID, &a.Variant, &a.AssignedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ExperimentAssignment{}, false, nil
		}
		return domain.ExperimentAssignment{}, false, err
	}
	a.AssignedAt = a.AssignedAt.UTC()
	return a, true, nil
}

type ReportRepo struct{ DB *sql.DB }

func (r *ReportRepo) SaveDef(ctx context.Context, def domain.ReportDef) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO report_defs (
			id, tenant_id, name, kind, query_spec, schedule, format, active, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, kind=EXCLUDED.kind, query_spec=EXCLUDED.query_spec,
			schedule=EXCLUDED.schedule, format=EXCLUDED.format, active=EXCLUDED.active,
			updated_at=EXCLUDED.updated_at`,
		def.ID, def.TenantID, def.Name, def.Kind, JSONMap(def.QuerySpec), def.Schedule, def.Format,
		def.Active, def.UpdatedAt.UTC())
	return err
}

func (r *ReportRepo) ListDefs(ctx context.Context, tenantID uuid.UUID) ([]domain.ReportDef, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, kind, query_spec, schedule, format, active, updated_at
		FROM report_defs WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ReportDef{}
	for rows.Next() {
		var def domain.ReportDef
		var spec JSONMap
		if err := rows.Scan(&def.ID, &def.TenantID, &def.Name, &def.Kind, &spec, &def.Schedule, &def.Format, &def.Active, &def.UpdatedAt); err != nil {
			return nil, err
		}
		def.QuerySpec = map[string]any(spec)
		def.UpdatedAt = def.UpdatedAt.UTC()
		out = append(out, def)
	}
	return out, rows.Err()
}

func (r *ReportRepo) SaveRun(ctx context.Context, run domain.ReportRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO report_runs (
			id, tenant_id, report_id, status, location, row_count, created_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			status=EXCLUDED.status, location=EXCLUDED.location, row_count=EXCLUDED.row_count,
			completed_at=EXCLUDED.completed_at`,
		run.ID, run.TenantID, run.ReportID, run.Status, run.Location, run.RowCount,
		run.CreatedAt.UTC(), nullTime(run.CompletedAt))
	return err
}

func (r *ReportRepo) ListRuns(ctx context.Context, tenantID, reportID uuid.UUID, limit int) ([]domain.ReportRun, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, report_id, status, location, row_count, created_at, completed_at
		FROM report_runs WHERE tenant_id=$1 AND report_id=$2
		ORDER BY created_at DESC LIMIT $3`, tenantID, reportID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ReportRun{}
	for rows.Next() {
		var run domain.ReportRun
		var completed sql.NullTime
		if err := rows.Scan(&run.ID, &run.TenantID, &run.ReportID, &run.Status, &run.Location, &run.RowCount, &run.CreatedAt, &completed); err != nil {
			return nil, err
		}
		run.CompletedAt = scanNullTime(completed)
		run.CreatedAt = run.CreatedAt.UTC()
		out = append(out, run)
	}
	return out, rows.Err()
}

type ObsRepo struct{ DB *sql.DB }

func (r *ObsRepo) SaveSignal(ctx context.Context, s domain.ObservabilitySignal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO observability_signals (
			id, tenant_id, kind, service, name, value, status, trace_id, attrs, occurred_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		s.ID, s.TenantID, s.Kind, s.Service, s.Name, s.Value, s.Status, s.TraceID,
		StringMap(s.Attrs), s.OccurredAt.UTC())
	return err
}

func (r *ObsRepo) ListSignals(ctx context.Context, tenantID uuid.UUID, kind, service string, limit int) ([]domain.ObservabilitySignal, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `
		SELECT id, tenant_id, kind, service, name, value, status, trace_id, attrs, occurred_at
		FROM observability_signals WHERE tenant_id=$1`
	args := []any{tenantID}
	n := 2
	if kind != "" {
		q += ` AND kind=$` + itoa(n)
		args = append(args, kind)
		n++
	}
	if service != "" {
		q += ` AND service=$` + itoa(n)
		args = append(args, service)
		n++
	}
	q += ` ORDER BY occurred_at DESC LIMIT $` + itoa(n)
	args = append(args, limit)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ObservabilitySignal{}
	for rows.Next() {
		var s domain.ObservabilitySignal
		var attrs StringMap
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Kind, &s.Service, &s.Name, &s.Value, &s.Status, &s.TraceID, &attrs, &s.OccurredAt); err != nil {
			return nil, err
		}
		s.Attrs = map[string]string(attrs)
		s.OccurredAt = s.OccurredAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	d := []byte{}
	for x := n; x > 0; x /= 10 {
		d = append([]byte{byte('0' + x%10)}, d...)
	}
	return string(d)
}

type AlertRepo struct{ DB *sql.DB }

func (r *AlertRepo) SaveRule(ctx context.Context, rule domain.AlertRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO alert_rules (
			id, tenant_id, name, metric_key, op, threshold, severity, enabled, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, metric_key=EXCLUDED.metric_key, op=EXCLUDED.op,
			threshold=EXCLUDED.threshold, severity=EXCLUDED.severity, enabled=EXCLUDED.enabled,
			updated_at=EXCLUDED.updated_at`,
		rule.ID, rule.TenantID, rule.Name, rule.MetricKey, rule.Op, rule.Threshold, rule.Severity,
		rule.Enabled, rule.UpdatedAt.UTC())
	return err
}

func (r *AlertRepo) ListRules(ctx context.Context, tenantID uuid.UUID) ([]domain.AlertRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, metric_key, op, threshold, severity, enabled, updated_at
		FROM alert_rules WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AlertRule{}
	for rows.Next() {
		var rule domain.AlertRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.MetricKey, &rule.Op, &rule.Threshold, &rule.Severity, &rule.Enabled, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rule.UpdatedAt = rule.UpdatedAt.UTC()
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *AlertRepo) SaveEvent(ctx context.Context, e domain.AlertEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO alert_events (
			id, tenant_id, rule_id, metric_key, value, severity, message, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.RuleID, e.MetricKey, e.Value, e.Severity, e.Message, e.CreatedAt.UTC())
	return err
}

func (r *AlertRepo) ListEvents(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AlertEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, rule_id, metric_key, value, severity, message, created_at
		FROM alert_events WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AlertEvent{}
	for rows.Next() {
		var e domain.AlertEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.RuleID, &e.MetricKey, &e.Value, &e.Severity, &e.Message, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type CatalogRepo struct{ DB *sql.DB }

func (r *CatalogRepo) SaveAsset(ctx context.Context, a domain.CatalogAsset) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO catalog_assets (
			id, tenant_id, name, type, owner, description, tags, classification, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, type=EXCLUDED.type, owner=EXCLUDED.owner, description=EXCLUDED.description,
			tags=EXCLUDED.tags, classification=EXCLUDED.classification, updated_at=EXCLUDED.updated_at`,
		a.ID, a.TenantID, a.Name, a.Type, a.Owner, a.Description, textArray(a.Tags), a.Classification, a.UpdatedAt.UTC())
	return err
}

func (r *CatalogRepo) ListAssets(ctx context.Context, tenantID uuid.UUID, typ string) ([]domain.CatalogAsset, error) {
	q := `
		SELECT id, tenant_id, name, type, owner, description, tags, classification, updated_at
		FROM catalog_assets WHERE tenant_id=$1`
	args := []any{tenantID}
	if typ != "" {
		q += ` AND type=$2`
		args = append(args, typ)
	}
	q += ` ORDER BY name ASC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CatalogAsset{}
	for rows.Next() {
		var a domain.CatalogAsset
		var tags []string
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &a.Type, &a.Owner, &a.Description, pq.Array(&tags), &a.Classification, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Tags = tags
		a.UpdatedAt = a.UpdatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *CatalogRepo) SaveLineage(ctx context.Context, e domain.LineageEdge) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO lineage_edges (tenant_id, from_name, to_name, kind) VALUES ($1,$2,$3,$4)`,
		e.TenantID, e.FromName, e.ToName, e.Kind)
	return err
}

func (r *CatalogRepo) ListLineage(ctx context.Context, tenantID uuid.UUID, name string) ([]domain.LineageEdge, error) {
	q := `SELECT tenant_id, from_name, to_name, kind FROM lineage_edges WHERE tenant_id=$1`
	args := []any{tenantID}
	if name != "" {
		q += ` AND (from_name=$2 OR to_name=$2)`
		args = append(args, name)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LineageEdge{}
	for rows.Next() {
		var e domain.LineageEdge
		if err := rows.Scan(&e.TenantID, &e.FromName, &e.ToName, &e.Kind); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type QualityRepo struct{ DB *sql.DB }

func (r *QualityRepo) Save(ctx context.Context, q domain.QualityCheck) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO quality_checks (
			id, tenant_id, asset_name, check_type, passed, score, details, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		q.ID, q.TenantID, q.AssetName, q.CheckType, q.Passed, q.Score, q.Details, q.CreatedAt.UTC())
	return err
}

func (r *QualityRepo) List(ctx context.Context, tenantID uuid.UUID, asset string, limit int) ([]domain.QualityCheck, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `
		SELECT id, tenant_id, asset_name, check_type, passed, score, details, created_at
		FROM quality_checks WHERE tenant_id=$1`
	args := []any{tenantID}
	n := 2
	if asset != "" {
		q += ` AND asset_name=$2`
		args = append(args, asset)
		n = 3
	}
	q += ` ORDER BY created_at DESC LIMIT $` + itoa(n)
	args = append(args, limit)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.QualityCheck{}
	for rows.Next() {
		var c domain.QualityCheck
		if err := rows.Scan(&c.ID, &c.TenantID, &c.AssetName, &c.CheckType, &c.Passed, &c.Score, &c.Details, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

var (
	_ ports.ExperimentRepo = (*ExperimentRepo)(nil)
	_ ports.ReportRepo     = (*ReportRepo)(nil)
	_ ports.ObsRepo        = (*ObsRepo)(nil)
	_ ports.AlertRepo      = (*AlertRepo)(nil)
	_ ports.CatalogRepo    = (*CatalogRepo)(nil)
	_ ports.QualityRepo    = (*QualityRepo)(nil)
)
