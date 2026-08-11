package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/app/ports"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

type KPIRepo struct{ DB *sql.DB }

func (r *KPIRepo) Save(ctx context.Context, k domain.KPI) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_kpis (id, tenant_id, key, name, value, target, period, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			key=EXCLUDED.key, name=EXCLUDED.name, value=EXCLUDED.value, target=EXCLUDED.target,
			period=EXCLUDED.period, updated_at=EXCLUDED.updated_at`,
		k.ID, k.TenantID, k.Key, k.Name, k.Value, k.Target, k.Period, k.UpdatedAt.UTC())
	return err
}

func (r *KPIRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.KPI, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, value, target, period, updated_at
		FROM eo_kpis WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.KPI{}
	for rows.Next() {
		var k domain.KPI
		if err := rows.Scan(&k.ID, &k.TenantID, &k.Key, &k.Name, &k.Value, &k.Target, &k.Period, &k.UpdatedAt); err != nil {
			return nil, err
		}
		k.UpdatedAt = k.UpdatedAt.UTC()
		out = append(out, k)
	}
	return out, rows.Err()
}

type RiskRepo struct{ DB *sql.DB }

func (r *RiskRepo) Save(ctx context.Context, risk domain.Risk) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_risks (
			id, tenant_id, code, title, category, likelihood, impact, score, status, owner_ref, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, title=EXCLUDED.title, category=EXCLUDED.category, likelihood=EXCLUDED.likelihood,
			impact=EXCLUDED.impact, score=EXCLUDED.score, status=EXCLUDED.status, owner_ref=EXCLUDED.owner_ref,
			updated_at=EXCLUDED.updated_at`,
		risk.ID, risk.TenantID, risk.Code, risk.Title, risk.Category, risk.Likelihood, risk.Impact,
		risk.Score, risk.Status, risk.OwnerRef, risk.CreatedAt.UTC(), risk.UpdatedAt.UTC())
	return err
}

func (r *RiskRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Risk, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, title, category, likelihood, impact, score, status, owner_ref, created_at, updated_at
		FROM eo_risks WHERE tenant_id=$1 ORDER BY score DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Risk{}
	for rows.Next() {
		var risk domain.Risk
		if err := rows.Scan(&risk.ID, &risk.TenantID, &risk.Code, &risk.Title, &risk.Category, &risk.Likelihood,
			&risk.Impact, &risk.Score, &risk.Status, &risk.OwnerRef, &risk.CreatedAt, &risk.UpdatedAt); err != nil {
			return nil, err
		}
		risk.CreatedAt = risk.CreatedAt.UTC()
		risk.UpdatedAt = risk.UpdatedAt.UTC()
		out = append(out, risk)
	}
	return out, rows.Err()
}

type ContinuityRepo struct{ DB *sql.DB }

func (r *ContinuityRepo) Save(ctx context.Context, p domain.ContinuityPlan) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_continuity (
			id, tenant_id, key, name, rto_hours, rpo_hours, priority, critical_service, status, activated_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, name=EXCLUDED.name, rto_hours=EXCLUDED.rto_hours, rpo_hours=EXCLUDED.rpo_hours,
			priority=EXCLUDED.priority, critical_service=EXCLUDED.critical_service, status=EXCLUDED.status,
			activated_at=EXCLUDED.activated_at, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Key, p.Name, p.RTOHours, p.RPOHours, p.Priority, p.CriticalSvc, p.Status,
		nullTime(p.ActivatedAt), p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *ContinuityRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.ContinuityPlan, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, rto_hours, rpo_hours, priority, critical_service, status, activated_at, created_at, updated_at
		FROM eo_continuity WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	return scanContinuity(row)
}

func (r *ContinuityRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ContinuityPlan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, rto_hours, rpo_hours, priority, critical_service, status, activated_at, created_at, updated_at
		FROM eo_continuity WHERE tenant_id=$1 ORDER BY priority ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ContinuityPlan{}
	for rows.Next() {
		p, err := scanContinuity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanContinuity(row scannable) (domain.ContinuityPlan, error) {
	var p domain.ContinuityPlan
	var activated sql.NullTime
	err := row.Scan(&p.ID, &p.TenantID, &p.Key, &p.Name, &p.RTOHours, &p.RPOHours, &p.Priority, &p.CriticalSvc,
		&p.Status, &activated, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ContinuityPlan{}, domain.ErrNotFound
		}
		return domain.ContinuityPlan{}, err
	}
	p.ActivatedAt = scanNullTime(activated)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

type AuditRepo struct{ DB *sql.DB }

func (r *AuditRepo) Save(ctx context.Context, a domain.AuditEngagement) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_audits (
			id, tenant_id, code, title, kind, status, scheduled_at, completed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, title=EXCLUDED.title, kind=EXCLUDED.kind, status=EXCLUDED.status,
			scheduled_at=EXCLUDED.scheduled_at, completed_at=EXCLUDED.completed_at`,
		a.ID, a.TenantID, a.Code, a.Title, a.Kind, a.Status, a.ScheduledAt.UTC(), nullTime(a.CompletedAt), a.CreatedAt.UTC())
	return err
}

func (r *AuditRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.AuditEngagement, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, kind, status, scheduled_at, completed_at, created_at
		FROM eo_audits WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.AuditEngagement
	var completed sql.NullTime
	err := row.Scan(&a.ID, &a.TenantID, &a.Code, &a.Title, &a.Kind, &a.Status, &a.ScheduledAt, &completed, &a.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.AuditEngagement{}, domain.ErrNotFound
		}
		return domain.AuditEngagement{}, err
	}
	a.CompletedAt = scanNullTime(completed)
	a.ScheduledAt = a.ScheduledAt.UTC()
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

func (r *AuditRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.AuditEngagement, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, title, kind, status, scheduled_at, completed_at, created_at
		FROM eo_audits WHERE tenant_id=$1 ORDER BY scheduled_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditEngagement{}
	for rows.Next() {
		var a domain.AuditEngagement
		var completed sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Code, &a.Title, &a.Kind, &a.Status, &a.ScheduledAt, &completed, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.CompletedAt = scanNullTime(completed)
		a.ScheduledAt = a.ScheduledAt.UTC()
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type FindingRepo struct{ DB *sql.DB }

func (r *FindingRepo) Save(ctx context.Context, f domain.AuditFinding) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_findings (id, tenant_id, audit_id, severity, title, capa, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			audit_id=EXCLUDED.audit_id, severity=EXCLUDED.severity, title=EXCLUDED.title,
			capa=EXCLUDED.capa, status=EXCLUDED.status`,
		f.ID, f.TenantID, f.AuditID, f.Severity, f.Title, f.CAPA, f.Status, f.CreatedAt.UTC())
	return err
}

func (r *FindingRepo) ListByAudit(ctx context.Context, tenantID, auditID uuid.UUID) ([]domain.AuditFinding, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, audit_id, severity, title, capa, status, created_at
		FROM eo_findings WHERE tenant_id=$1 AND audit_id=$2 ORDER BY created_at DESC`, tenantID, auditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditFinding{}
	for rows.Next() {
		var f domain.AuditFinding
		if err := rows.Scan(&f.ID, &f.TenantID, &f.AuditID, &f.Severity, &f.Title, &f.CAPA, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.CreatedAt = f.CreatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

type MeetingRepo struct{ DB *sql.DB }

func (r *MeetingRepo) Save(ctx context.Context, m domain.Meeting) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_meetings (id, tenant_id, kind, title, starts_at, agenda, minutes_uri, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			kind=EXCLUDED.kind, title=EXCLUDED.title, starts_at=EXCLUDED.starts_at, agenda=EXCLUDED.agenda,
			minutes_uri=EXCLUDED.minutes_uri, status=EXCLUDED.status`,
		m.ID, m.TenantID, m.Kind, m.Title, m.StartsAt.UTC(), m.Agenda, m.MinutesURI, m.Status, m.CreatedAt.UTC())
	return err
}

func (r *MeetingRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Meeting, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, title, starts_at, agenda, minutes_uri, status, created_at
		FROM eo_meetings WHERE tenant_id=$1 ORDER BY starts_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Meeting{}
	for rows.Next() {
		var m domain.Meeting
		if err := rows.Scan(&m.ID, &m.TenantID, &m.Kind, &m.Title, &m.StartsAt, &m.Agenda, &m.MinutesURI, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.StartsAt = m.StartsAt.UTC()
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

type DecisionRepo struct{ DB *sql.DB }

func (r *DecisionRepo) Save(ctx context.Context, d domain.Decision) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_decisions (
			id, tenant_id, title, body, meeting_id, decided_by, votes_for, votes_against, impact, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			title=EXCLUDED.title, body=EXCLUDED.body, meeting_id=EXCLUDED.meeting_id, decided_by=EXCLUDED.decided_by,
			votes_for=EXCLUDED.votes_for, votes_against=EXCLUDED.votes_against, impact=EXCLUDED.impact`,
		d.ID, d.TenantID, d.Title, d.Body, nullUUID(d.MeetingID), d.DecidedBy, d.VotesFor, d.VotesAgainst,
		d.Impact, d.CreatedAt.UTC())
	return err
}

func (r *DecisionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Decision, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, title, body, meeting_id, decided_by, votes_for, votes_against, impact, created_at
		FROM eo_decisions WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Decision{}
	for rows.Next() {
		var d domain.Decision
		var meeting uuid.NullUUID
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Title, &d.Body, &meeting, &d.DecidedBy, &d.VotesFor,
			&d.VotesAgainst, &d.Impact, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.MeetingID = scanNullUUID(meeting)
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

type KnowledgeRepo struct{ DB *sql.DB }

func (r *KnowledgeRepo) Save(ctx context.Context, d domain.KnowledgeDoc) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_knowledge (id, tenant_id, key, title, kind, uri, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET key=EXCLUDED.key, title=EXCLUDED.title, kind=EXCLUDED.kind, uri=EXCLUDED.uri`,
		d.ID, d.TenantID, d.Key, d.Title, d.Kind, d.URI, d.CreatedAt.UTC())
	return err
}

func (r *KnowledgeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.KnowledgeDoc, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, title, kind, uri, created_at
		FROM eo_knowledge WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.KnowledgeDoc{}
	for rows.Next() {
		var d domain.KnowledgeDoc
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Key, &d.Title, &d.Kind, &d.URI, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

type ResourceRepo struct{ DB *sql.DB }

func (r *ResourceRepo) Save(ctx context.Context, plan domain.ResourcePlan) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eo_resources (
			id, tenant_id, team_code, period, capacity_fte, allocated_fte, utilization, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			team_code=EXCLUDED.team_code, period=EXCLUDED.period, capacity_fte=EXCLUDED.capacity_fte,
			allocated_fte=EXCLUDED.allocated_fte, utilization=EXCLUDED.utilization, updated_at=EXCLUDED.updated_at`,
		plan.ID, plan.TenantID, plan.TeamCode, plan.Period, plan.CapacityFTE, plan.AllocatedFTE,
		plan.Utilization, plan.UpdatedAt.UTC())
	return err
}

func (r *ResourceRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ResourcePlan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, team_code, period, capacity_fte, allocated_fte, utilization, updated_at
		FROM eo_resources WHERE tenant_id=$1 ORDER BY team_code ASC, period DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ResourcePlan{}
	for rows.Next() {
		var plan domain.ResourcePlan
		if err := rows.Scan(&plan.ID, &plan.TenantID, &plan.TeamCode, &plan.Period, &plan.CapacityFTE,
			&plan.AllocatedFTE, &plan.Utilization, &plan.UpdatedAt); err != nil {
			return nil, err
		}
		plan.UpdatedAt = plan.UpdatedAt.UTC()
		out = append(out, plan)
	}
	return out, rows.Err()
}

var (
	_ ports.KPIRepo        = (*KPIRepo)(nil)
	_ ports.RiskRepo       = (*RiskRepo)(nil)
	_ ports.ContinuityRepo = (*ContinuityRepo)(nil)
	_ ports.AuditRepo      = (*AuditRepo)(nil)
	_ ports.FindingRepo    = (*FindingRepo)(nil)
	_ ports.MeetingRepo    = (*MeetingRepo)(nil)
	_ ports.DecisionRepo   = (*DecisionRepo)(nil)
	_ ports.KnowledgeRepo  = (*KnowledgeRepo)(nil)
	_ ports.ResourceRepo   = (*ResourceRepo)(nil)
)
