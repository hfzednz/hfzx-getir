package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/domain"
)

type PolicyRepo struct{ DB *sql.DB }

func (r *PolicyRepo) Save(ctx context.Context, p domain.SecurityPolicy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_policies (id, tenant_id, key, kind, version, rego, enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, key) DO UPDATE SET id=EXCLUDED.id, kind=EXCLUDED.kind, version=EXCLUDED.version,
			rego=EXCLUDED.rego, enabled=EXCLUDED.enabled, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Key, p.Kind, p.Version, p.Rego, p.Enabled, p.CreatedAt.UTC(), p.UpdatedAt.UTC())
	return err
}

func (r *PolicyRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SecurityPolicy, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, kind, version, rego, enabled, created_at, updated_at
		FROM sec_policies WHERE tenant_id=$1 AND key=$2`, tenantID, domain.NormalizeKey(key))
	var p domain.SecurityPolicy
	err := row.Scan(&p.ID, &p.TenantID, &p.Key, &p.Kind, &p.Version, &p.Rego, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.SecurityPolicy{}, domain.ErrNotFound
		}
		return domain.SecurityPolicy{}, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p, nil
}

func (r *PolicyRepo) List(ctx context.Context, tenantID uuid.UUID, kind string) ([]domain.SecurityPolicy, error) {
	q := `
		SELECT id, tenant_id, key, kind, version, rego, enabled, created_at, updated_at
		FROM sec_policies WHERE tenant_id=$1`
	args := []any{tenantID}
	if kind != "" {
		q += ` AND kind=$2`
		args = append(args, kind)
	}
	q += ` ORDER BY key ASC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SecurityPolicy{}
	for rows.Next() {
		var p domain.SecurityPolicy
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Key, &p.Kind, &p.Version, &p.Rego, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		p.UpdatedAt = p.UpdatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PolicyRepo) SaveDecision(ctx context.Context, d domain.PolicyDecision) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_policy_decisions (
			id, tenant_id, policy_key, subject, action, resource, allow, reason, risk_score, context_json, evaluated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		d.ID, d.TenantID, d.PolicyKey, d.Subject, d.Action, d.Resource, d.Allow, d.Reason, d.RiskScore, JSONMap(d.Context), d.EvaluatedAt.UTC())
	return err
}

type AuditRepo struct{ DB *sql.DB }

func (r *AuditRepo) Append(ctx context.Context, e domain.AuditEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_audit_events (
			id, tenant_id, actor_id, actor_type, action, resource_type, resource_id, outcome,
			ip, user_agent, metadata_json, hash, prev_hash, occurred_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		e.ID, e.TenantID, e.ActorID, e.ActorType, e.Action, e.ResourceType, e.ResourceID, e.Outcome,
		e.IP, e.UserAgent, JSONMap(e.Metadata), e.Hash, e.PrevHash, e.OccurredAt.UTC())
	return err
}

func (r *AuditRepo) LastHash(ctx context.Context, tenantID uuid.UUID) (string, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT hash FROM sec_audit_events WHERE tenant_id=$1 ORDER BY occurred_at DESC LIMIT 1`, tenantID)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if isNoRows(err) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

func (r *AuditRepo) Search(ctx context.Context, tenantID uuid.UUID, action, actor string, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `
		SELECT id, tenant_id, actor_id, actor_type, action, resource_type, resource_id, outcome,
			ip, user_agent, metadata_json, hash, prev_hash, occurred_at
		FROM sec_audit_events WHERE tenant_id=$1`
	args := []any{tenantID}
	n := 2
	if action != "" {
		q += fmt.Sprintf(` AND action=$%d`, n)
		args = append(args, action)
		n++
	}
	if actor != "" {
		q += fmt.Sprintf(` AND actor_id=$%d`, n)
		args = append(args, actor)
		n++
	}
	q += fmt.Sprintf(` ORDER BY occurred_at DESC LIMIT $%d`, n)
	args = append(args, limit)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditEvent{}
	for rows.Next() {
		var e domain.AuditEvent
		var meta JSONMap
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ActorID, &e.ActorType, &e.Action, &e.ResourceType, &e.ResourceID, &e.Outcome,
			&e.IP, &e.UserAgent, &meta, &e.Hash, &e.PrevHash, &e.OccurredAt); err != nil {
			return nil, err
		}
		e.Metadata = map[string]any(meta)
		e.OccurredAt = e.OccurredAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type SecretRepo struct{ DB *sql.DB }

func (r *SecretRepo) Save(ctx context.Context, s domain.SecretMeta) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_secrets (
			id, tenant_id, name, kind, vault_path, version, rotatable, expires_at, last_rotated, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, name) DO UPDATE SET id=EXCLUDED.id, kind=EXCLUDED.kind, vault_path=EXCLUDED.vault_path,
			version=EXCLUDED.version, rotatable=EXCLUDED.rotatable, expires_at=EXCLUDED.expires_at,
			last_rotated=EXCLUDED.last_rotated, status=EXCLUDED.status`,
		s.ID, s.TenantID, s.Name, s.Kind, s.VaultPath, s.Version, s.Rotatable, nullTime(s.ExpiresAt), nullTime(s.LastRotated), s.Status, s.CreatedAt.UTC())
	return err
}

func (r *SecretRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SecretMeta, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, kind, vault_path, version, rotatable, expires_at, last_rotated, status, created_at
		FROM sec_secrets WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanSecret(row)
}

func (r *SecretRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.SecretMeta, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, kind, vault_path, version, rotatable, expires_at, last_rotated, status, created_at
		FROM sec_secrets WHERE tenant_id=$1 ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SecretMeta{}
	for rows.Next() {
		s, err := scanSecret(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanSecret(row scannable) (domain.SecretMeta, error) {
	var s domain.SecretMeta
	var exp, rot sql.NullTime
	err := row.Scan(&s.ID, &s.TenantID, &s.Name, &s.Kind, &s.VaultPath, &s.Version, &s.Rotatable, &exp, &rot, &s.Status, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.SecretMeta{}, domain.ErrNotFound
		}
		return domain.SecretMeta{}, err
	}
	s.ExpiresAt = scanNullTime(exp)
	s.LastRotated = scanNullTime(rot)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

type ThreatRepo struct{ DB *sql.DB }

func (r *ThreatRepo) Save(ctx context.Context, t domain.ThreatAlert) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_threat_alerts (id, tenant_id, kind, severity, subject, score, indicators_json, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET kind=EXCLUDED.kind, severity=EXCLUDED.severity, subject=EXCLUDED.subject,
			score=EXCLUDED.score, indicators_json=EXCLUDED.indicators_json, status=EXCLUDED.status`,
		t.ID, t.TenantID, t.Kind, t.Severity, t.Subject, t.Score, JSONMap(t.Indicators), t.Status, t.CreatedAt.UTC())
	return err
}

func (r *ThreatRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ThreatAlert, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, severity, subject, score, indicators_json, status, created_at
		FROM sec_threat_alerts WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var t domain.ThreatAlert
	var ind JSONMap
	err := row.Scan(&t.ID, &t.TenantID, &t.Kind, &t.Severity, &t.Subject, &t.Score, &ind, &t.Status, &t.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ThreatAlert{}, domain.ErrNotFound
		}
		return domain.ThreatAlert{}, err
	}
	t.Indicators = map[string]any(ind)
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}

func (r *ThreatRepo) List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.ThreatAlert, error) {
	q := `
		SELECT id, tenant_id, kind, severity, subject, score, indicators_json, status, created_at
		FROM sec_threat_alerts WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ThreatAlert{}
	for rows.Next() {
		var t domain.ThreatAlert
		var ind JSONMap
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Kind, &t.Severity, &t.Subject, &t.Score, &ind, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Indicators = map[string]any(ind)
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type VulnRepo struct{ DB *sql.DB }

func (r *VulnRepo) Save(ctx context.Context, f domain.ScanFinding) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_scan_findings (id, tenant_id, source, target, cve, severity, title, status, detected_at, fixed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET source=EXCLUDED.source, target=EXCLUDED.target, cve=EXCLUDED.cve,
			severity=EXCLUDED.severity, title=EXCLUDED.title, status=EXCLUDED.status, fixed_at=EXCLUDED.fixed_at`,
		f.ID, f.TenantID, f.Source, f.Target, f.CVE, f.Severity, f.Title, f.Status, f.DetectedAt.UTC(), nullTime(f.FixedAt))
	return err
}

func (r *VulnRepo) List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.ScanFinding, error) {
	q := `
		SELECT id, tenant_id, source, target, cve, severity, title, status, detected_at, fixed_at
		FROM sec_scan_findings WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY detected_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ScanFinding{}
	for rows.Next() {
		var f domain.ScanFinding
		var fixed sql.NullTime
		if err := rows.Scan(&f.ID, &f.TenantID, &f.Source, &f.Target, &f.CVE, &f.Severity, &f.Title, &f.Status, &f.DetectedAt, &fixed); err != nil {
			return nil, err
		}
		f.FixedAt = scanNullTime(fixed)
		f.DetectedAt = f.DetectedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

type IncidentRepo struct{ DB *sql.DB }

func (r *IncidentRepo) Save(ctx context.Context, i domain.Incident) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sec_incidents (
			id, tenant_id, title, severity, status, threat_id, timeline_json, playbook_key, assignee, opened_at, closed_at, postmortem
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, severity=EXCLUDED.severity, status=EXCLUDED.status,
			threat_id=EXCLUDED.threat_id, timeline_json=EXCLUDED.timeline_json, playbook_key=EXCLUDED.playbook_key,
			assignee=EXCLUDED.assignee, closed_at=EXCLUDED.closed_at, postmortem=EXCLUDED.postmortem`,
		i.ID, i.TenantID, i.Title, i.Severity, i.Status, nullUUID(i.ThreatID), JSONRaw{V: i.Timeline}, i.PlaybookKey, i.Assignee,
		i.OpenedAt.UTC(), nullTime(i.ClosedAt), i.Postmortem)
	return err
}

func (r *IncidentRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Incident, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, severity, status, threat_id, timeline_json, playbook_key, assignee, opened_at, closed_at, postmortem
		FROM sec_incidents WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanIncident(row)
}

func (r *IncidentRepo) List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.Incident, error) {
	q := `
		SELECT id, tenant_id, title, severity, status, threat_id, timeline_json, playbook_key, assignee, opened_at, closed_at, postmortem
		FROM sec_incidents WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY opened_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Incident{}
	for rows.Next() {
		i, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func scanIncident(row scannable) (domain.Incident, error) {
	var i domain.Incident
	var threat uuid.NullUUID
	var timelineRaw []byte
	var closed sql.NullTime
	err := row.Scan(&i.ID, &i.TenantID, &i.Title, &i.Severity, &i.Status, &threat, &timelineRaw, &i.PlaybookKey, &i.Assignee, &i.OpenedAt, &closed, &i.Postmortem)
	if err != nil {
		if isNoRows(err) {
			return domain.Incident{}, domain.ErrNotFound
		}
		return domain.Incident{}, err
	}
	i.ThreatID = scanNullUUID(threat)
	i.ClosedAt = scanNullTime(closed)
	i.OpenedAt = i.OpenedAt.UTC()
	if len(timelineRaw) > 0 {
		_ = json.Unmarshal(timelineRaw, &i.Timeline)
	}
	if i.Timeline == nil {
		i.Timeline = []domain.IncidentEvent{}
	}
	return i, nil
}
