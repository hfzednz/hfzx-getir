package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/app/ports"
	"github.com/nexora/security-service/internal/domain"
)

type ComplianceRepo struct{ DB *sql.DB }

func (r *ComplianceRepo) SaveControl(ctx context.Context, c domain.ComplianceControl) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_compliance_controls (id, tenant_id, framework, control_id, title, status, evidence_ids, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, framework, control_id) DO UPDATE SET id=EXCLUDED.id, title=EXCLUDED.title,
			status=EXCLUDED.status, evidence_ids=EXCLUDED.evidence_ids, updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.Framework, c.ControlID, c.Title, c.Status, UUIDArray(c.EvidenceIDs), c.UpdatedAt.UTC())
	return err
}

func (r *ComplianceRepo) ListControls(ctx context.Context, tenantID uuid.UUID, framework string) ([]domain.ComplianceControl, error) {
	q := `
		SELECT id, tenant_id, framework, control_id, title, status, evidence_ids, updated_at
		FROM sec_compliance_controls WHERE tenant_id=$1`
	args := []any{tenantID}
	if framework != "" {
		q += ` AND framework=$2`
		args = append(args, framework)
	}
	q += ` ORDER BY framework, control_id`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ComplianceControl{}
	for rows.Next() {
		var c domain.ComplianceControl
		var evid UUIDArray
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Framework, &c.ControlID, &c.Title, &c.Status, &evid, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.EvidenceIDs = []uuid.UUID(evid)
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ComplianceRepo) SaveEvidence(ctx context.Context, e domain.ComplianceEvidence) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_compliance_evidence (id, tenant_id, control_id, title, uri, collected_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, uri=EXCLUDED.uri, collected_at=EXCLUDED.collected_at`,
		e.ID, e.TenantID, e.ControlID, e.Title, e.URI, e.CollectedAt.UTC())
	return err
}

func (r *ComplianceRepo) SaveAuditRun(ctx context.Context, run domain.ComplianceAuditRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_compliance_audit_runs (id, tenant_id, framework, score, gaps, status, started_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET score=EXCLUDED.score, gaps=EXCLUDED.gaps, status=EXCLUDED.status, completed_at=EXCLUDED.completed_at`,
		run.ID, run.TenantID, run.Framework, run.Score, run.Gaps, run.Status, run.StartedAt.UTC(), nullTime(run.CompletedAt))
	return err
}

func (r *ComplianceRepo) ListAuditRuns(ctx context.Context, tenantID uuid.UUID) ([]domain.ComplianceAuditRun, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, framework, score, gaps, status, started_at, completed_at
		FROM sec_compliance_audit_runs WHERE tenant_id=$1 ORDER BY started_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ComplianceAuditRun{}
	for rows.Next() {
		var run domain.ComplianceAuditRun
		var completed sql.NullTime
		if err := rows.Scan(&run.ID, &run.TenantID, &run.Framework, &run.Score, &run.Gaps, &run.Status, &run.StartedAt, &completed); err != nil {
			return nil, err
		}
		run.CompletedAt = scanNullTime(completed)
		run.StartedAt = run.StartedAt.UTC()
		out = append(out, run)
	}
	return out, rows.Err()
}

type DataGovRepo struct{ DB *sql.DB }

func (r *DataGovRepo) SaveAsset(ctx context.Context, a domain.DataAsset) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_data_assets (id, tenant_id, name, classification, pii_tags, retention_days, owner, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, classification=EXCLUDED.classification,
			pii_tags=EXCLUDED.pii_tags, retention_days=EXCLUDED.retention_days, owner=EXCLUDED.owner`,
		a.ID, a.TenantID, a.Name, a.Classification, TextArray(a.PIITags), a.RetentionDays, a.Owner, a.CreatedAt.UTC())
	return err
}

func (r *DataGovRepo) ListAssets(ctx context.Context, tenantID uuid.UUID) ([]domain.DataAsset, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, classification, pii_tags, retention_days, owner, created_at
		FROM sec_data_assets WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DataAsset{}
	for rows.Next() {
		var a domain.DataAsset
		var tags TextArray
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &a.Classification, &tags, &a.RetentionDays, &a.Owner, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.PIITags = []string(tags)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *DataGovRepo) SavePrivacy(ctx context.Context, p domain.PrivacyRequest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_privacy_requests (id, tenant_id, subject_ref, kind, status, created_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, completed_at=EXCLUDED.completed_at`,
		p.ID, p.TenantID, p.SubjectRef, p.Kind, p.Status, p.CreatedAt.UTC(), nullTime(p.CompletedAt))
	return err
}

func (r *DataGovRepo) GetPrivacy(ctx context.Context, tenantID, id uuid.UUID) (domain.PrivacyRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, subject_ref, kind, status, created_at, completed_at
		FROM sec_privacy_requests WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.PrivacyRequest
	var completed sql.NullTime
	err := row.Scan(&p.ID, &p.TenantID, &p.SubjectRef, &p.Kind, &p.Status, &p.CreatedAt, &completed)
	if err != nil {
		if isNoRows(err) {
			return domain.PrivacyRequest{}, domain.ErrNotFound
		}
		return domain.PrivacyRequest{}, err
	}
	p.CompletedAt = scanNullTime(completed)
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *DataGovRepo) ListPrivacy(ctx context.Context, tenantID uuid.UUID) ([]domain.PrivacyRequest, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, subject_ref, kind, status, created_at, completed_at
		FROM sec_privacy_requests WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PrivacyRequest{}
	for rows.Next() {
		var p domain.PrivacyRequest
		var completed sql.NullTime
		if err := rows.Scan(&p.ID, &p.TenantID, &p.SubjectRef, &p.Kind, &p.Status, &p.CreatedAt, &completed); err != nil {
			return nil, err
		}
		p.CompletedAt = scanNullTime(completed)
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type RiskRepo struct{ DB *sql.DB }

func (r *RiskRepo) Save(ctx context.Context, item domain.RiskItem) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_risks (id, tenant_id, title, category, likelihood, impact, score, status, owner, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, category=EXCLUDED.category, likelihood=EXCLUDED.likelihood,
			impact=EXCLUDED.impact, score=EXCLUDED.score, status=EXCLUDED.status, owner=EXCLUDED.owner, updated_at=EXCLUDED.updated_at`,
		item.ID, item.TenantID, item.Title, item.Category, item.Likelihood, item.Impact, item.Score, item.Status, item.Owner,
		item.CreatedAt.UTC(), item.UpdatedAt.UTC())
	return err
}

func (r *RiskRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.RiskItem, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, category, likelihood, impact, score, status, owner, created_at, updated_at
		FROM sec_risks WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var item domain.RiskItem
	err := row.Scan(&item.ID, &item.TenantID, &item.Title, &item.Category, &item.Likelihood, &item.Impact, &item.Score,
		&item.Status, &item.Owner, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.RiskItem{}, domain.ErrNotFound
		}
		return domain.RiskItem{}, err
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

func (r *RiskRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RiskItem, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, title, category, likelihood, impact, score, status, owner, created_at, updated_at
		FROM sec_risks WHERE tenant_id=$1 ORDER BY score DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RiskItem{}
	for rows.Next() {
		var item domain.RiskItem
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Title, &item.Category, &item.Likelihood, &item.Impact, &item.Score,
			&item.Status, &item.Owner, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = item.CreatedAt.UTC()
		item.UpdatedAt = item.UpdatedAt.UTC()
		out = append(out, item)
	}
	return out, rows.Err()
}

type AccessRepo struct{ DB *sql.DB }

func (r *AccessRepo) Save(ctx context.Context, a domain.AccessRequest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_access_requests (
			id, tenant_id, requester_id, role_hint, resource, reason, ttl_minutes, status, created_at, decided_at, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, decided_at=EXCLUDED.decided_at, expires_at=EXCLUDED.expires_at`,
		a.ID, a.TenantID, a.RequesterID, a.RoleHint, a.Resource, a.Reason, a.TTLMinutes, a.Status,
		a.CreatedAt.UTC(), nullTime(a.DecidedAt), nullTime(a.ExpiresAt))
	return err
}

func (r *AccessRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.AccessRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, requester_id, role_hint, resource, reason, ttl_minutes, status, created_at, decided_at, expires_at
		FROM sec_access_requests WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanAccess(row)
}

func (r *AccessRepo) ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.AccessRequest, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, requester_id, role_hint, resource, reason, ttl_minutes, status, created_at, decided_at, expires_at
		FROM sec_access_requests WHERE tenant_id=$1 AND status='pending' ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AccessRequest{}
	for rows.Next() {
		a, err := scanAccess(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAccess(row scannable) (domain.AccessRequest, error) {
	var a domain.AccessRequest
	var decided, expires sql.NullTime
	err := row.Scan(&a.ID, &a.TenantID, &a.RequesterID, &a.RoleHint, &a.Resource, &a.Reason, &a.TTLMinutes, &a.Status,
		&a.CreatedAt, &decided, &expires)
	if err != nil {
		if isNoRows(err) {
			return domain.AccessRequest{}, domain.ErrNotFound
		}
		return domain.AccessRequest{}, err
	}
	a.DecidedAt = scanNullTime(decided)
	a.ExpiresAt = scanNullTime(expires)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

type DeviceRepo struct{ DB *sql.DB }

func (r *DeviceRepo) Save(ctx context.Context, d domain.DeviceTrust) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_devices (
			id, tenant_id, device_id, platform, attested, rooted, jailbroken, tampered, trust_score, last_seen_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (tenant_id, device_id) DO UPDATE SET id=EXCLUDED.id, platform=EXCLUDED.platform, attested=EXCLUDED.attested,
			rooted=EXCLUDED.rooted, jailbroken=EXCLUDED.jailbroken, tampered=EXCLUDED.tampered,
			trust_score=EXCLUDED.trust_score, last_seen_at=EXCLUDED.last_seen_at`,
		d.ID, d.TenantID, d.DeviceID, d.Platform, d.Attested, d.Rooted, d.Jailbroken, d.Tampered, d.TrustScore, d.LastSeenAt.UTC())
	return err
}

func (r *DeviceRepo) GetByDevice(ctx context.Context, tenantID uuid.UUID, deviceID string) (domain.DeviceTrust, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, device_id, platform, attested, rooted, jailbroken, tampered, trust_score, last_seen_at
		FROM sec_devices WHERE tenant_id=$1 AND device_id=$2`, tenantID, deviceID)
	var d domain.DeviceTrust
	err := row.Scan(&d.ID, &d.TenantID, &d.DeviceID, &d.Platform, &d.Attested, &d.Rooted, &d.Jailbroken, &d.Tampered, &d.TrustScore, &d.LastSeenAt)
	if err != nil {
		if isNoRows(err) {
			return domain.DeviceTrust{}, domain.ErrNotFound
		}
		return domain.DeviceTrust{}, err
	}
	d.LastSeenAt = d.LastSeenAt.UTC()
	return d, nil
}

type AISecRepo struct{ DB *sql.DB }

func (r *AISecRepo) Save(ctx context.Context, e domain.AISecurityEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_ai_events (id, tenant_id, model_key, prompt_hash, kind, blocked, score, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.ModelKey, e.PromptHash, e.Kind, e.Blocked, e.Score, e.CreatedAt.UTC())
	return err
}

func (r *AISecRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.AISecurityEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, model_key, prompt_hash, kind, blocked, score, created_at
		FROM sec_ai_events WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AISecurityEvent{}
	for rows.Next() {
		var e domain.AISecurityEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ModelKey, &e.PromptHash, &e.Kind, &e.Blocked, &e.Score, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type FraudRepo struct{ DB *sql.DB }

func (r *FraudRepo) Save(ctx context.Context, s domain.FraudSignal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_fraud_signals (id, tenant_id, subject, kind, score, features_json, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		s.ID, s.TenantID, s.Subject, s.Kind, s.Score, FloatMap(s.Features), s.CreatedAt.UTC())
	return err
}

var (
	_ ports.PolicyRepo     = (*PolicyRepo)(nil)
	_ ports.AuditRepo      = (*AuditRepo)(nil)
	_ ports.SecretRepo     = (*SecretRepo)(nil)
	_ ports.ThreatRepo     = (*ThreatRepo)(nil)
	_ ports.VulnRepo       = (*VulnRepo)(nil)
	_ ports.IncidentRepo   = (*IncidentRepo)(nil)
	_ ports.ComplianceRepo = (*ComplianceRepo)(nil)
	_ ports.DataGovRepo    = (*DataGovRepo)(nil)
	_ ports.RiskRepo       = (*RiskRepo)(nil)
	_ ports.AccessRepo     = (*AccessRepo)(nil)
	_ ports.DeviceRepo     = (*DeviceRepo)(nil)
	_ ports.AISecRepo      = (*AISecRepo)(nil)
	_ ports.FraudRepo      = (*FraudRepo)(nil)
)
