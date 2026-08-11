package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/app/ports"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
)

type AuditRepo struct{ DB *sql.DB }

func (r *AuditRepo) Save(ctx context.Context, a domain.Audit) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_audits (id, tenant_id, domain, title, status, created_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, status=EXCLUDED.status, completed_at=EXCLUDED.completed_at`,
		a.ID, a.TenantID, string(a.Domain), a.Title, a.Status, a.CreatedAt.UTC(), nullTime(a.CompletedAt))
	return err
}

func (r *AuditRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Audit, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, domain, title, status, created_at, completed_at
		FROM hs_audits WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.Audit
	var dom string
	var completed sql.NullTime
	err := row.Scan(&a.ID, &a.TenantID, &dom, &a.Title, &a.Status, &a.CreatedAt, &completed)
	if err != nil {
		if isNoRows(err) {
			return domain.Audit{}, domain.ErrNotFound
		}
		return domain.Audit{}, err
	}
	a.Domain = domain.AuditDomain(dom)
	a.CompletedAt = scanNullTime(completed)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

func (r *AuditRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Audit, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, domain, title, status, created_at, completed_at
		FROM hs_audits WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Audit{}
	for rows.Next() {
		var a domain.Audit
		var dom string
		var completed sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &dom, &a.Title, &a.Status, &a.CreatedAt, &completed); err != nil {
			return nil, err
		}
		a.Domain = domain.AuditDomain(dom)
		a.CompletedAt = scanNullTime(completed)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type FindingRepo struct{ DB *sql.DB }

func (r *FindingRepo) Save(ctx context.Context, f domain.Finding) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_findings (id, tenant_id, audit_id, code, title, severity, status, resolution, created_at, resolved_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, resolution=EXCLUDED.resolution, resolved_at=EXCLUDED.resolved_at`,
		f.ID, f.TenantID, f.AuditID, f.Code, f.Title, string(f.Severity), f.Status, f.Resolution, f.CreatedAt.UTC(), nullTime(f.ResolvedAt))
	return err
}

func (r *FindingRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Finding, error) {
	return r.queryFindings(ctx, `SELECT id, tenant_id, audit_id, code, title, severity, status, resolution, created_at, resolved_at
		FROM hs_findings WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
}

func (r *FindingRepo) ListByAudit(ctx context.Context, tenantID, auditID uuid.UUID) ([]domain.Finding, error) {
	return r.queryFindings(ctx, `SELECT id, tenant_id, audit_id, code, title, severity, status, resolution, created_at, resolved_at
		FROM hs_findings WHERE tenant_id=$1 AND audit_id=$2 ORDER BY created_at DESC`, tenantID, auditID)
}

func (r *FindingRepo) OpenCritical(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hs_findings
		WHERE tenant_id=$1 AND status='open' AND severity=$2`, tenantID, string(domain.SeverityCritical)).Scan(&n)
	return n, err
}

func (r *FindingRepo) queryFindings(ctx context.Context, q string, args ...any) ([]domain.Finding, error) {
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Finding{}
	for rows.Next() {
		var f domain.Finding
		var sev string
		var resolved sql.NullTime
		if err := rows.Scan(&f.ID, &f.TenantID, &f.AuditID, &f.Code, &f.Title, &sev, &f.Status, &f.Resolution, &f.CreatedAt, &resolved); err != nil {
			return nil, err
		}
		f.Severity = domain.FindingSeverity(sev)
		f.ResolvedAt = scanNullTime(resolved)
		f.CreatedAt = f.CreatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

type BenchmarkRepo struct{ DB *sql.DB }

func (r *BenchmarkRepo) Save(ctx context.Context, b domain.BenchmarkRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_benchmarks (id, tenant_id, kind, value, target, passed, scenario, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.TenantID, string(b.Kind), b.Value, b.Target, b.Passed, b.Scenario, b.CreatedAt.UTC())
	return err
}

func (r *BenchmarkRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.BenchmarkRun, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, value, target, passed, scenario, created_at
		FROM hs_benchmarks WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBenchmarks(rows)
}

func (r *BenchmarkRepo) LatestByKind(ctx context.Context, tenantID uuid.UUID, kind domain.BenchmarkKind) (domain.BenchmarkRun, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, value, target, passed, scenario, created_at
		FROM hs_benchmarks WHERE tenant_id=$1 AND kind=$2 ORDER BY created_at DESC LIMIT 1`, tenantID, string(kind))
	var b domain.BenchmarkRun
	var k string
	err := row.Scan(&b.ID, &b.TenantID, &k, &b.Value, &b.Target, &b.Passed, &b.Scenario, &b.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.BenchmarkRun{}, domain.ErrNotFound
		}
		return domain.BenchmarkRun{}, err
	}
	b.Kind = domain.BenchmarkKind(k)
	b.CreatedAt = b.CreatedAt.UTC()
	return b, nil
}

func scanBenchmarks(rows *sql.Rows) ([]domain.BenchmarkRun, error) {
	out := []domain.BenchmarkRun{}
	for rows.Next() {
		var b domain.BenchmarkRun
		var k string
		if err := rows.Scan(&b.ID, &b.TenantID, &k, &b.Value, &b.Target, &b.Passed, &b.Scenario, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.Kind = domain.BenchmarkKind(k)
		b.CreatedAt = b.CreatedAt.UTC()
		out = append(out, b)
	}
	return out, rows.Err()
}

type CapacityRepo struct{ DB *sql.DB }

func (r *CapacityRepo) Save(ctx context.Context, c domain.CapacityScenario) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_capacity (id, tenant_id, key, name, peak_rps, regions, notes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, peak_rps=EXCLUDED.peak_rps, regions=EXCLUDED.regions, notes=EXCLUDED.notes`,
		c.ID, c.TenantID, c.Key, c.Name, c.PeakRPS, TextArray(c.Regions), c.Notes, c.CreatedAt.UTC())
	return err
}

func (r *CapacityRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CapacityScenario, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, peak_rps, regions, notes, created_at
		FROM hs_capacity WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CapacityScenario{}
	for rows.Next() {
		var c domain.CapacityScenario
		var regions TextArray
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Key, &c.Name, &c.PeakRPS, &regions, &c.Notes, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Regions = []string(regions)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type ChaosRepo struct{ DB *sql.DB }

func (r *ChaosRepo) Save(ctx context.Context, c domain.ChaosExperiment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_chaos (id, tenant_id, kind, name, status, recovery_sec, created_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, recovery_sec=EXCLUDED.recovery_sec, completed_at=EXCLUDED.completed_at`,
		c.ID, c.TenantID, string(c.Kind), c.Name, c.Status, c.RecoverySec, c.CreatedAt.UTC(), nullTime(c.CompletedAt))
	return err
}

func (r *ChaosRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChaosExperiment, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, name, status, recovery_sec, created_at, completed_at
		FROM hs_chaos WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.ChaosExperiment
	var kind string
	var completed sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &kind, &c.Name, &c.Status, &c.RecoverySec, &c.CreatedAt, &completed)
	if err != nil {
		if isNoRows(err) {
			return domain.ChaosExperiment{}, domain.ErrNotFound
		}
		return domain.ChaosExperiment{}, err
	}
	c.Kind = domain.ChaosKind(kind)
	c.CompletedAt = scanNullTime(completed)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *ChaosRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.ChaosExperiment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, name, status, recovery_sec, created_at, completed_at
		FROM hs_chaos WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ChaosExperiment{}
	for rows.Next() {
		var c domain.ChaosExperiment
		var kind string
		var completed sql.NullTime
		if err := rows.Scan(&c.ID, &c.TenantID, &kind, &c.Name, &c.Status, &c.RecoverySec, &c.CreatedAt, &completed); err != nil {
			return nil, err
		}
		c.Kind = domain.ChaosKind(kind)
		c.CompletedAt = scanNullTime(completed)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type TuningRepo struct{ DB *sql.DB }

func (r *TuningRepo) Save(ctx context.Context, t domain.TuningProfile) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_tuning (id, tenant_id, key, layer, uri, applied, created_at, applied_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, layer=EXCLUDED.layer, uri=EXCLUDED.uri, applied=EXCLUDED.applied, applied_at=EXCLUDED.applied_at`,
		t.ID, t.TenantID, t.Key, t.Layer, t.URI, t.Applied, t.CreatedAt.UTC(), nullTime(t.AppliedAt))
	return err
}

func (r *TuningRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.TuningProfile, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, layer, uri, applied, created_at, applied_at
		FROM hs_tuning WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var t domain.TuningProfile
	var appliedAt sql.NullTime
	err := row.Scan(&t.ID, &t.TenantID, &t.Key, &t.Layer, &t.URI, &t.Applied, &t.CreatedAt, &appliedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.TuningProfile{}, domain.ErrNotFound
		}
		return domain.TuningProfile{}, err
	}
	t.AppliedAt = scanNullTime(appliedAt)
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}

func (r *TuningRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.TuningProfile, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, layer, uri, applied, created_at, applied_at
		FROM hs_tuning WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TuningProfile{}
	for rows.Next() {
		var t domain.TuningProfile
		var appliedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Key, &t.Layer, &t.URI, &t.Applied, &t.CreatedAt, &appliedAt); err != nil {
			return nil, err
		}
		t.AppliedAt = scanNullTime(appliedAt)
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type CertificateRepo struct{ DB *sql.DB }

func (r *CertificateRepo) Save(ctx context.Context, c domain.Certificate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO hs_certificates (id, tenant_id, version, status, gates, issued_at, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, gates=EXCLUDED.gates, issued_at=EXCLUDED.issued_at, expires_at=EXCLUDED.expires_at`,
		c.ID, c.TenantID, c.Version, c.Status, BoolMap(c.Gates), nullTime(c.IssuedAt), nullTime(c.ExpiresAt), c.CreatedAt.UTC())
	return err
}

func (r *CertificateRepo) Latest(ctx context.Context, tenantID uuid.UUID) (domain.Certificate, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, version, status, gates, issued_at, expires_at, created_at
		FROM hs_certificates WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 1`, tenantID)
	c, err := scanCert(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Certificate{}, domain.ErrNotFound
		}
		return domain.Certificate{}, err
	}
	return c, nil
}

func (r *CertificateRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Certificate, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, version, status, gates, issued_at, expires_at, created_at
		FROM hs_certificates WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Certificate{}
	for rows.Next() {
		c, err := scanCert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanCert(row scannable) (domain.Certificate, error) {
	var c domain.Certificate
	var gates BoolMap
	var issued, expires sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &c.Version, &c.Status, &gates, &issued, &expires, &c.CreatedAt)
	if err != nil {
		return domain.Certificate{}, err
	}
	c.Gates = map[string]bool(gates)
	c.IssuedAt = scanNullTime(issued)
	c.ExpiresAt = scanNullTime(expires)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

var (
	_ ports.AuditRepo       = (*AuditRepo)(nil)
	_ ports.FindingRepo     = (*FindingRepo)(nil)
	_ ports.BenchmarkRepo   = (*BenchmarkRepo)(nil)
	_ ports.CapacityRepo    = (*CapacityRepo)(nil)
	_ ports.ChaosRepo       = (*ChaosRepo)(nil)
	_ ports.TuningRepo      = (*TuningRepo)(nil)
	_ ports.CertificateRepo = (*CertificateRepo)(nil)
)
