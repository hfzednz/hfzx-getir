package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/app/ports"
	"github.com/nexora/ai-platform-service/internal/domain"
)

// FeatureRepo persists feature records.
type FeatureRepo struct{ DB *sql.DB }

func (r *FeatureRepo) Upsert(ctx context.Context, f domain.FeatureRecord) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO feature_records (
			tenant_id, entity_type, entity_id, name, version, "values", tags, lineage, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, entity_type, entity_id, name, version) DO UPDATE SET
			"values"=EXCLUDED."values", tags=EXCLUDED.tags, lineage=EXCLUDED.lineage, updated_at=EXCLUDED.updated_at`,
		f.TenantID, f.EntityType, f.EntityID, f.Name, f.Version,
		JSONFloatMap(f.Values), JSONStringMap(f.Tags), f.Lineage, f.UpdatedAt.UTC())
	return err
}

func (r *FeatureRepo) Get(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, name string, version int) (domain.FeatureRecord, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, entity_type, entity_id, name, version, "values", tags, lineage, updated_at
		FROM feature_records
		WHERE tenant_id=$1 AND entity_type=$2 AND entity_id=$3 AND name=$4 AND version=$5`,
		tenantID, entityType, entityID, name, version)
	var f domain.FeatureRecord
	var values JSONFloatMap
	var tags JSONStringMap
	err := row.Scan(&f.TenantID, &f.EntityType, &f.EntityID, &f.Name, &f.Version, &values, &tags, &f.Lineage, &f.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.FeatureRecord{}, domain.ErrNotFound
		}
		return domain.FeatureRecord{}, err
	}
	f.Values = map[string]float64(values)
	f.Tags = map[string]string(tags)
	f.UpdatedAt = f.UpdatedAt.UTC()
	return f, nil
}

func (r *FeatureRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]domain.FeatureRecord, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, entity_type, entity_id, name, version, "values", tags, lineage, updated_at
		FROM feature_records WHERE tenant_id=$1 AND entity_type=$2 AND entity_id=$3
		ORDER BY name, version`, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FeatureRecord{}
	for rows.Next() {
		var f domain.FeatureRecord
		var values JSONFloatMap
		var tags JSONStringMap
		if err := rows.Scan(&f.TenantID, &f.EntityType, &f.EntityID, &f.Name, &f.Version, &values, &tags, &f.Lineage, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.Values = map[string]float64(values)
		f.Tags = map[string]string(tags)
		f.UpdatedAt = f.UpdatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

// ModelRepo persists model cards.
type ModelRepo struct{ DB *sql.DB }

func (r *ModelRepo) Save(ctx context.Context, m domain.ModelCard) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO model_cards (
			id, tenant_id, key, name, framework, version, stage, artifact_uri, metrics,
			approved_by, approved_at, deploy_strat, canary_pct, shadow, fallback_key, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, name=EXCLUDED.name, framework=EXCLUDED.framework, version=EXCLUDED.version,
			stage=EXCLUDED.stage, artifact_uri=EXCLUDED.artifact_uri, metrics=EXCLUDED.metrics,
			approved_by=EXCLUDED.approved_by, approved_at=EXCLUDED.approved_at, deploy_strat=EXCLUDED.deploy_strat,
			canary_pct=EXCLUDED.canary_pct, shadow=EXCLUDED.shadow, fallback_key=EXCLUDED.fallback_key,
			updated_at=EXCLUDED.updated_at`,
		m.ID, m.TenantID, m.Key, m.Name, m.Framework, m.Version, m.Stage, m.ArtifactURI, JSONFloatMap(m.Metrics),
		nullUUID(m.ApprovedBy), nullTime(m.ApprovedAt), m.DeployStrat, m.CanaryPct, m.Shadow, m.FallbackKey,
		m.CreatedAt.UTC(), m.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *ModelRepo) Get(ctx context.Context, tenantID uuid.UUID, key, version string) (domain.ModelCard, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, framework, version, stage, artifact_uri, metrics,
			approved_by, approved_at, deploy_strat, canary_pct, shadow, fallback_key, created_at, updated_at
		FROM model_cards WHERE tenant_id=$1 AND key=$2 AND version=$3`, tenantID, key, version)
	return scanModel(row)
}

func (r *ModelRepo) GetProd(ctx context.Context, tenantID uuid.UUID, key string) (domain.ModelCard, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, framework, version, stage, artifact_uri, metrics,
			approved_by, approved_at, deploy_strat, canary_pct, shadow, fallback_key, created_at, updated_at
		FROM model_cards WHERE tenant_id=$1 AND key=$2 AND stage=$3
		ORDER BY updated_at DESC LIMIT 1`, tenantID, key, domain.StageProd)
	return scanModel(row)
}

func (r *ModelRepo) List(ctx context.Context, tenantID uuid.UUID, key string) ([]domain.ModelCard, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, framework, version, stage, artifact_uri, metrics,
			approved_by, approved_at, deploy_strat, canary_pct, shadow, fallback_key, created_at, updated_at
		FROM model_cards WHERE tenant_id=$1 AND ($2='' OR key=$2)
		ORDER BY key, version`, tenantID, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ModelCard{}
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanModel(row scannable) (domain.ModelCard, error) {
	var m domain.ModelCard
	var metrics JSONFloatMap
	var approvedBy uuid.NullUUID
	var approvedAt sql.NullTime
	err := row.Scan(
		&m.ID, &m.TenantID, &m.Key, &m.Name, &m.Framework, &m.Version, &m.Stage, &m.ArtifactURI, &metrics,
		&approvedBy, &approvedAt, &m.DeployStrat, &m.CanaryPct, &m.Shadow, &m.FallbackKey, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ModelCard{}, domain.ErrNotFound
		}
		return domain.ModelCard{}, err
	}
	m.Metrics = map[string]float64(metrics)
	m.ApprovedBy = scanNullUUID(approvedBy)
	m.ApprovedAt = scanNullTime(approvedAt)
	m.CreatedAt = m.CreatedAt.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

// PromptRepo persists prompt templates.
type PromptRepo struct{ DB *sql.DB }

func (r *PromptRepo) Save(ctx context.Context, p domain.PromptTemplate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO prompt_templates (
			id, tenant_id, key, locale, system, user_tpl, version, active, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, locale=EXCLUDED.locale, system=EXCLUDED.system, user_tpl=EXCLUDED.user_tpl,
			version=EXCLUDED.version, active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Key, p.Locale, p.System, p.UserTpl, p.Version, p.Active, p.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *PromptRepo) GetActive(ctx context.Context, tenantID uuid.UUID, key, locale string) (domain.PromptTemplate, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, locale, system, user_tpl, version, active, updated_at
		FROM prompt_templates
		WHERE tenant_id=$1 AND key=$2 AND active=TRUE AND ($3='' OR locale='' OR locale=$3)
		ORDER BY CASE WHEN locale=$3 THEN 0 ELSE 1 END, updated_at DESC LIMIT 1`,
		tenantID, key, locale)
	var p domain.PromptTemplate
	err := row.Scan(&p.ID, &p.TenantID, &p.Key, &p.Locale, &p.System, &p.UserTpl, &p.Version, &p.Active, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.PromptTemplate{}, domain.ErrNotFound
		}
		return domain.PromptTemplate{}, err
	}
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *PromptRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.PromptTemplate, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, locale, system, user_tpl, version, active, updated_at
		FROM prompt_templates WHERE tenant_id=$1 ORDER BY key, version`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PromptTemplate{}
	for rows.Next() {
		var p domain.PromptTemplate
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Key, &p.Locale, &p.System, &p.UserTpl, &p.Version, &p.Active, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

// MemoryRepo persists conversation memory.
type MemoryRepo struct{ DB *sql.DB }

func (r *MemoryRepo) Append(ctx context.Context, m domain.ConversationMemory) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO conversation_memory (id, tenant_id, session_id, role, content, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.TenantID, m.SessionID, m.Role, m.Content, m.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *MemoryRepo) ListSession(ctx context.Context, tenantID, sessionID uuid.UUID, limit int) ([]domain.ConversationMemory, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, session_id, role, content, created_at
		FROM (
			SELECT id, tenant_id, session_id, role, content, created_at
			FROM conversation_memory WHERE tenant_id=$1 AND session_id=$2
			ORDER BY created_at DESC LIMIT $3
		) t ORDER BY created_at ASC`, tenantID, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ConversationMemory{}
	for rows.Next() {
		var m domain.ConversationMemory
		if err := rows.Scan(&m.ID, &m.TenantID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

// AgentRepo persists agent runs.
type AgentRepo struct{ DB *sql.DB }

func (r *AgentRepo) SaveRun(ctx context.Context, run domain.AgentRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO agent_runs (id, tenant_id, kind, input, output, steps, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, input=EXCLUDED.input, output=EXCLUDED.output,
			steps=EXCLUDED.steps, status=EXCLUDED.status`,
		run.ID, run.TenantID, run.Kind, run.Input, run.Output, JSONSteps(run.Steps), run.Status, run.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AgentRepo) GetRun(ctx context.Context, tenantID, id uuid.UUID) (domain.AgentRun, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, input, output, steps, status, created_at
		FROM agent_runs WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	var run domain.AgentRun
	var steps JSONSteps
	err := row.Scan(&run.ID, &run.TenantID, &run.Kind, &run.Input, &run.Output, &steps, &run.Status, &run.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.AgentRun{}, domain.ErrNotFound
		}
		return domain.AgentRun{}, err
	}
	run.Steps = []domain.AgentStep(steps)
	run.CreatedAt = run.CreatedAt.UTC()
	return run, nil
}

func (r *AgentRepo) ListRuns(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.AgentRun, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, input, output, steps, status, created_at
		FROM agent_runs WHERE tenant_id=$1 AND ($2='' OR kind=$2)
		ORDER BY created_at DESC LIMIT $3`, tenantID, kind, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AgentRun{}
	for rows.Next() {
		var run domain.AgentRun
		var steps JSONSteps
		if err := rows.Scan(&run.ID, &run.TenantID, &run.Kind, &run.Input, &run.Output, &steps, &run.Status, &run.CreatedAt); err != nil {
			return nil, err
		}
		run.Steps = []domain.AgentStep(steps)
		run.CreatedAt = run.CreatedAt.UTC()
		out = append(out, run)
	}
	return out, rows.Err()
}

// AutomationRepo persists automation rules and runs.
type AutomationRepo struct{ DB *sql.DB }

func (r *AutomationRepo) SaveRule(ctx context.Context, rule domain.AutomationRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO automation_rules (
			id, tenant_id, name, enabled, priority, conditions, actions, require_approval, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, enabled=EXCLUDED.enabled, priority=EXCLUDED.priority,
			conditions=EXCLUDED.conditions, actions=EXCLUDED.actions,
			require_approval=EXCLUDED.require_approval, updated_at=EXCLUDED.updated_at`,
		rule.ID, rule.TenantID, rule.Name, rule.Enabled, rule.Priority,
		JSONConditions(rule.Conditions), JSONActions(rule.Actions), rule.RequireApproval, rule.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AutomationRepo) ListRules(ctx context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, enabled, priority, conditions, actions, require_approval, updated_at
		FROM automation_rules WHERE tenant_id=$1 ORDER BY priority DESC, name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AutomationRule{}
	for rows.Next() {
		var rule domain.AutomationRule
		var conds JSONConditions
		var acts JSONActions
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Enabled, &rule.Priority,
			&conds, &acts, &rule.RequireApproval, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rule.Conditions = []domain.RuleCondition(conds)
		rule.Actions = []domain.RuleAction(acts)
		rule.UpdatedAt = rule.UpdatedAt.UTC()
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *AutomationRepo) SaveRun(ctx context.Context, run domain.AutomationRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO automation_runs (id, tenant_id, rule_id, matched, approved, result, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		run.ID, run.TenantID, run.RuleID, run.Matched, run.Approved, run.Result, run.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

// DriftRepo persists drift reports.
type DriftRepo struct{ DB *sql.DB }

func (r *DriftRepo) Save(ctx context.Context, d domain.DriftReport) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO drift_reports (id, tenant_id, model_key, metric, value, threshold, severity, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		d.ID, d.TenantID, d.ModelKey, d.Metric, d.Value, d.Threshold, d.Severity, d.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *DriftRepo) List(ctx context.Context, tenantID uuid.UUID, modelKey string, limit int) ([]domain.DriftReport, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, model_key, metric, value, threshold, severity, created_at
		FROM drift_reports WHERE tenant_id=$1 AND ($2='' OR model_key=$2)
		ORDER BY created_at DESC LIMIT $3`, tenantID, modelKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DriftReport{}
	for rows.Next() {
		var d domain.DriftReport
		if err := rows.Scan(&d.ID, &d.TenantID, &d.ModelKey, &d.Metric, &d.Value, &d.Threshold, &d.Severity, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

var (
	_ ports.FeatureRepo    = (*FeatureRepo)(nil)
	_ ports.ModelRepo      = (*ModelRepo)(nil)
	_ ports.PromptRepo     = (*PromptRepo)(nil)
	_ ports.MemoryRepo     = (*MemoryRepo)(nil)
	_ ports.AgentRepo      = (*AgentRepo)(nil)
	_ ports.AutomationRepo = (*AutomationRepo)(nil)
	_ ports.DriftRepo      = (*DriftRepo)(nil)
)
