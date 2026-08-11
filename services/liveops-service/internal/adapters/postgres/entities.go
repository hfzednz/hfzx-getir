package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/app/ports"
	"github.com/nexora/liveops-service/internal/domain"
)

// FlagRepo persists feature flags.
type FlagRepo struct{ DB *sql.DB }

func (r *FlagRepo) Save(ctx context.Context, f domain.FeatureFlag) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_flags (
			id, tenant_id, key, description, enabled, percentage, rules_json, depends_on,
			version, emergency_off, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, description=EXCLUDED.description, enabled=EXCLUDED.enabled,
			percentage=EXCLUDED.percentage, rules_json=EXCLUDED.rules_json, depends_on=EXCLUDED.depends_on,
			version=EXCLUDED.version, emergency_off=EXCLUDED.emergency_off, updated_at=EXCLUDED.updated_at`,
		f.ID, f.TenantID, f.Key, f.Description, f.Enabled, f.Percentage, JSONRules(f.Rules), TextArray(f.DependsOn),
		f.Version, f.EmergencyOff, f.CreatedAt.UTC(), f.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *FlagRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.FeatureFlag, error) {
	key = domain.NormalizeKey(key)
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, description, enabled, percentage, rules_json, depends_on,
			version, emergency_off, created_at, updated_at
		FROM liveops_flags WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	f, err := scanFlag(row)
	if err != nil {
		if isNoRows(err) {
			return domain.FeatureFlag{}, domain.ErrNotFound
		}
		return domain.FeatureFlag{}, err
	}
	return f, nil
}

func (r *FlagRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.FeatureFlag, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, description, enabled, percentage, rules_json, depends_on,
			version, emergency_off, created_at, updated_at
		FROM liveops_flags WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FeatureFlag{}
	for rows.Next() {
		f, err := scanFlag(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanFlag(row scannable) (domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	var rules JSONRules
	var deps TextArray
	err := row.Scan(
		&f.ID, &f.TenantID, &f.Key, &f.Description, &f.Enabled, &f.Percentage, &rules, &deps,
		&f.Version, &f.EmergencyOff, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return domain.FeatureFlag{}, err
	}
	f.Rules = []domain.TargetRule(rules)
	f.DependsOn = []string(deps)
	f.CreatedAt = f.CreatedAt.UTC()
	f.UpdatedAt = f.UpdatedAt.UTC()
	return f, nil
}

// ConfigRepo persists remote config documents.
type ConfigRepo struct{ DB *sql.DB }

func (r *ConfigRepo) Save(ctx context.Context, c domain.ConfigDocument) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_configs (
			id, tenant_id, key, namespace, payload, version, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, namespace=EXCLUDED.namespace, payload=EXCLUDED.payload,
			version=EXCLUDED.version, status=EXCLUDED.status, updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.Key, c.Namespace, JSONMap(c.Payload), c.Version, c.Status,
		c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ConfigRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.ConfigDocument, error) {
	key = domain.NormalizeKey(key)
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, namespace, payload, version, status, created_at, updated_at
		FROM liveops_configs WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var c domain.ConfigDocument
	var payload JSONMap
	err := row.Scan(&c.ID, &c.TenantID, &c.Key, &c.Namespace, &payload, &c.Version, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ConfigDocument{}, domain.ErrNotFound
		}
		return domain.ConfigDocument{}, err
	}
	c.Payload = map[string]any(payload)
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func (r *ConfigRepo) List(ctx context.Context, tenantID uuid.UUID, namespace string) ([]domain.ConfigDocument, error) {
	q := `
		SELECT id, tenant_id, key, namespace, payload, version, status, created_at, updated_at
		FROM liveops_configs WHERE tenant_id=$1`
	args := []any{tenantID}
	if namespace != "" {
		q += ` AND namespace=$2`
		args = append(args, namespace)
	}
	q += ` ORDER BY key ASC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ConfigDocument{}
	for rows.Next() {
		var c domain.ConfigDocument
		var payload JSONMap
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Key, &c.Namespace, &payload, &c.Version, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Payload = map[string]any(payload)
		c.CreatedAt = c.CreatedAt.UTC()
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

// ExperimentRepo persists experiments and assignments.
type ExperimentRepo struct{ DB *sql.DB }

func (r *ExperimentRepo) Save(ctx context.Context, e domain.Experiment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_experiments (
			id, tenant_id, key, name, status, kind, hypothesis, variants_json, primary_metric,
			started_at, ended_at, winner, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, name=EXCLUDED.name, status=EXCLUDED.status, kind=EXCLUDED.kind,
			hypothesis=EXCLUDED.hypothesis, variants_json=EXCLUDED.variants_json,
			primary_metric=EXCLUDED.primary_metric, started_at=EXCLUDED.started_at,
			ended_at=EXCLUDED.ended_at, winner=EXCLUDED.winner`,
		e.ID, e.TenantID, e.Key, e.Name, e.Status, e.Kind, e.Hypothesis, JSONVariants(e.Variants), e.PrimaryMetric,
		nullTime(e.StartedAt), nullTime(e.EndedAt), e.Winner, e.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ExperimentRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error) {
	key = domain.NormalizeKey(key)
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, status, kind, hypothesis, variants_json, primary_metric,
			started_at, ended_at, winner, created_at
		FROM liveops_experiments WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	e, err := scanExperiment(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Experiment{}, domain.ErrNotFound
		}
		return domain.Experiment{}, err
	}
	return e, nil
}

func (r *ExperimentRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Experiment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, status, kind, hypothesis, variants_json, primary_metric,
			started_at, ended_at, winner, created_at
		FROM liveops_experiments WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Experiment{}
	for rows.Next() {
		e, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ExperimentRepo) SaveAssignment(ctx context.Context, a domain.Assignment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_assignments (
			id, tenant_id, experiment_id, subject_id, variant_key, assigned_at
		) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, experiment_id, subject_id) DO UPDATE SET
			variant_key=EXCLUDED.variant_key, assigned_at=EXCLUDED.assigned_at, id=EXCLUDED.id`,
		a.ID, a.TenantID, a.ExperimentID, a.SubjectID, a.VariantKey, a.AssignedAt.UTC())
	return err
}

func (r *ExperimentRepo) GetAssignment(ctx context.Context, tenantID, experimentID uuid.UUID, subjectID string) (domain.Assignment, bool, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, experiment_id, subject_id, variant_key, assigned_at
		FROM liveops_assignments WHERE tenant_id=$1 AND experiment_id=$2 AND subject_id=$3`,
		tenantID, experimentID, subjectID)
	var a domain.Assignment
	err := row.Scan(&a.ID, &a.TenantID, &a.ExperimentID, &a.SubjectID, &a.VariantKey, &a.AssignedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Assignment{}, false, nil
		}
		return domain.Assignment{}, false, err
	}
	a.AssignedAt = a.AssignedAt.UTC()
	return a, true, nil
}

func scanExperiment(row scannable) (domain.Experiment, error) {
	var e domain.Experiment
	var variants JSONVariants
	var started, ended sql.NullTime
	err := row.Scan(
		&e.ID, &e.TenantID, &e.Key, &e.Name, &e.Status, &e.Kind, &e.Hypothesis, &variants, &e.PrimaryMetric,
		&started, &ended, &e.Winner, &e.CreatedAt)
	if err != nil {
		return domain.Experiment{}, err
	}
	e.Variants = []domain.Variant(variants)
	e.StartedAt = scanNullTime(started)
	e.EndedAt = scanNullTime(ended)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

// EventRepo persists liveops calendar events.
type EventRepo struct{ DB *sql.DB }

func (r *EventRepo) Save(ctx context.Context, e domain.LiveOpsEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_events (
			id, tenant_id, key, kind, title, status, starts_at, ends_at, scopes_json, config_patch, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, kind=EXCLUDED.kind, title=EXCLUDED.title, status=EXCLUDED.status,
			starts_at=EXCLUDED.starts_at, ends_at=EXCLUDED.ends_at, scopes_json=EXCLUDED.scopes_json,
			config_patch=EXCLUDED.config_patch`,
		e.ID, e.TenantID, e.Key, e.Kind, e.Title, e.Status, e.StartsAt.UTC(), e.EndsAt.UTC(),
		JSONRules(e.Scopes), JSONMap(e.ConfigPatch), e.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *EventRepo) List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.LiveOpsEvent, error) {
	q := `
		SELECT id, tenant_id, key, kind, title, status, starts_at, ends_at, scopes_json, config_patch, created_at
		FROM liveops_events WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY starts_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LiveOpsEvent{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EventRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.LiveOpsEvent, error) {
	key = domain.NormalizeKey(key)
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, kind, title, status, starts_at, ends_at, scopes_json, config_patch, created_at
		FROM liveops_events WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	e, err := scanEvent(row)
	if err != nil {
		if isNoRows(err) {
			return domain.LiveOpsEvent{}, domain.ErrNotFound
		}
		return domain.LiveOpsEvent{}, err
	}
	return e, nil
}

func scanEvent(row scannable) (domain.LiveOpsEvent, error) {
	var e domain.LiveOpsEvent
	var scopes JSONRules
	var patch JSONMap
	err := row.Scan(
		&e.ID, &e.TenantID, &e.Key, &e.Kind, &e.Title, &e.Status, &e.StartsAt, &e.EndsAt, &scopes, &patch, &e.CreatedAt)
	if err != nil {
		return domain.LiveOpsEvent{}, err
	}
	e.Scopes = []domain.TargetRule(scopes)
	e.ConfigPatch = map[string]any(patch)
	e.StartsAt = e.StartsAt.UTC()
	e.EndsAt = e.EndsAt.UTC()
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

// ChangeRepo persists change requests.
type ChangeRepo struct{ DB *sql.DB }

func (r *ChangeRepo) Save(ctx context.Context, c domain.ChangeRequest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_changes (
			id, tenant_id, kind, subject_key, payload, status, created_at, decided_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, subject_key=EXCLUDED.subject_key, payload=EXCLUDED.payload,
			status=EXCLUDED.status, decided_at=EXCLUDED.decided_at`,
		c.ID, c.TenantID, c.Kind, c.SubjectKey, JSONMap(c.Payload), c.Status, c.CreatedAt.UTC(), nullTime(c.DecidedAt))
	return err
}

func (r *ChangeRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, subject_key, payload, status, created_at, decided_at
		FROM liveops_changes WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.ChangeRequest
	var payload JSONMap
	var decided sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &c.Kind, &c.SubjectKey, &payload, &c.Status, &c.CreatedAt, &decided)
	if err != nil {
		if isNoRows(err) {
			return domain.ChangeRequest{}, domain.ErrNotFound
		}
		return domain.ChangeRequest{}, err
	}
	c.Payload = map[string]any(payload)
	c.DecidedAt = scanNullTime(decided)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *ChangeRepo) ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.ChangeRequest, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, subject_key, payload, status, created_at, decided_at
		FROM liveops_changes WHERE tenant_id=$1 AND status='pending' ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ChangeRequest{}
	for rows.Next() {
		var c domain.ChangeRequest
		var payload JSONMap
		var decided sql.NullTime
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Kind, &c.SubjectKey, &payload, &c.Status, &c.CreatedAt, &decided); err != nil {
			return nil, err
		}
		c.Payload = map[string]any(payload)
		c.DecidedAt = scanNullTime(decided)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

// RollbackRepo persists rollback records.
type RollbackRepo struct{ DB *sql.DB }

func (r *RollbackRepo) Save(ctx context.Context, rec domain.RollbackRecord) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO liveops_rollbacks (
			id, tenant_id, kind, subject_key, from_version, to_version, reason, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO NOTHING`,
		rec.ID, rec.TenantID, rec.Kind, rec.SubjectKey, rec.FromVersion, rec.ToVersion, rec.Reason, rec.CreatedAt.UTC())
	return err
}

func (r *RollbackRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RollbackRecord, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, subject_key, from_version, to_version, reason, created_at
		FROM liveops_rollbacks WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RollbackRecord{}
	for rows.Next() {
		var rec domain.RollbackRecord
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.Kind, &rec.SubjectKey, &rec.FromVersion, &rec.ToVersion, &rec.Reason, &rec.CreatedAt); err != nil {
			return nil, err
		}
		rec.CreatedAt = rec.CreatedAt.UTC()
		out = append(out, rec)
	}
	return out, rows.Err()
}

var (
	_ ports.FlagRepo       = (*FlagRepo)(nil)
	_ ports.ConfigRepo     = (*ConfigRepo)(nil)
	_ ports.ExperimentRepo = (*ExperimentRepo)(nil)
	_ ports.EventRepo      = (*EventRepo)(nil)
	_ ports.ChangeRepo     = (*ChangeRepo)(nil)
	_ ports.RollbackRepo   = (*RollbackRepo)(nil)
)