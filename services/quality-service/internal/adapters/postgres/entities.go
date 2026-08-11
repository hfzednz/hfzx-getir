package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/app/ports"
	"github.com/nexora/quality-service/internal/domain"
)

type SuiteRepo struct{ DB *sql.DB }

func (r *SuiteRepo) Save(ctx context.Context, s domain.Suite) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_suites (id, tenant_id, key, name, kind, owner, path, enabled, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, name=EXCLUDED.name, kind=EXCLUDED.kind, owner=EXCLUDED.owner,
			path=EXCLUDED.path, enabled=EXCLUDED.enabled`,
		s.ID, s.TenantID, s.Key, s.Name, string(s.Kind), s.Owner, s.Path, s.Enabled, s.CreatedAt.UTC())
	return err
}
func (r *SuiteRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Suite, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, kind, owner, path, enabled, created_at
		FROM qa_suites WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var s domain.Suite
	var kind string
	err := row.Scan(&s.ID, &s.TenantID, &s.Key, &s.Name, &kind, &s.Owner, &s.Path, &s.Enabled, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Suite{}, domain.ErrNotFound
		}
		return domain.Suite{}, err
	}
	s.Kind = domain.SuiteKind(kind)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}
func (r *SuiteRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Suite, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, name, kind, owner, path, enabled, created_at
		FROM qa_suites WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Suite{}
	for rows.Next() {
		var s domain.Suite
		var kind string
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Key, &s.Name, &kind, &s.Owner, &s.Path, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Kind = domain.SuiteKind(kind)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type RunRepo struct{ DB *sql.DB }

func (r *RunRepo) Save(ctx context.Context, run domain.TestRun) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_runs (
			id, tenant_id, suite_key, kind, status, trigger, commit_sha, branch, environment,
			started_at, finished_at, summary_json
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			status=EXCLUDED.status, finished_at=EXCLUDED.finished_at, summary_json=EXCLUDED.summary_json`,
		run.ID, run.TenantID, run.SuiteKey, string(run.Kind), string(run.Status), run.Trigger, run.CommitSHA, run.Branch, run.Environment,
		run.StartedAt.UTC(), nullTime(run.FinishedAt), SummaryJSON(run.Summary))
	return err
}
func (r *RunRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.TestRun, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, suite_key, kind, status, trigger, commit_sha, branch, environment,
			started_at, finished_at, summary_json
		FROM qa_runs WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	run, err := scanRun(row)
	if err != nil {
		if isNoRows(err) {
			return domain.TestRun{}, domain.ErrNotFound
		}
		return domain.TestRun{}, err
	}
	return run, nil
}
func (r *RunRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.TestRun, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, suite_key, kind, status, trigger, commit_sha, branch, environment,
			started_at, finished_at, summary_json
		FROM qa_runs WHERE tenant_id=$1 ORDER BY started_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TestRun{}
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanRun(row scannable) (domain.TestRun, error) {
	var run domain.TestRun
	var kind, status string
	var finished sql.NullTime
	var summary SummaryJSON
	err := row.Scan(&run.ID, &run.TenantID, &run.SuiteKey, &kind, &status, &run.Trigger, &run.CommitSHA, &run.Branch, &run.Environment,
		&run.StartedAt, &finished, &summary)
	if err != nil {
		return domain.TestRun{}, err
	}
	run.Kind = domain.SuiteKind(kind)
	run.Status = domain.RunStatus(status)
	run.FinishedAt = scanNullTime(finished)
	run.Summary = domain.RunSummary(summary)
	run.StartedAt = run.StartedAt.UTC()
	return run, nil
}

type ResultRepo struct{ DB *sql.DB }

func (r *ResultRepo) Save(ctx context.Context, res domain.TestCaseResult) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_results (id, tenant_id, run_id, name, class_name, status, duration_ms, message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, duration_ms=EXCLUDED.duration_ms, message=EXCLUDED.message`,
		res.ID, res.TenantID, res.RunID, res.Name, res.ClassName, res.Status, res.DurationMs, res.Message)
	return err
}
func (r *ResultRepo) ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.TestCaseResult, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, run_id, name, class_name, status, duration_ms, message
		FROM qa_results WHERE tenant_id=$1 AND run_id=$2`, tenantID, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TestCaseResult{}
	for rows.Next() {
		var res domain.TestCaseResult
		if err := rows.Scan(&res.ID, &res.TenantID, &res.RunID, &res.Name, &res.ClassName, &res.Status, &res.DurationMs, &res.Message); err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

type CoverageRepo struct{ DB *sql.DB }

func (r *CoverageRepo) Save(ctx context.Context, c domain.CoverageReport) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_coverage (id, tenant_id, run_id, service, line_pct, branch_pct, api_pct, workflow_pct, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.TenantID, c.RunID, c.Service, c.LinePct, c.BranchPct, c.APIPct, c.WorkflowPct, c.CreatedAt.UTC())
	return err
}
func (r *CoverageRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CoverageReport, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, run_id, service, line_pct, branch_pct, api_pct, workflow_pct, created_at
		FROM qa_coverage WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CoverageReport{}
	for rows.Next() {
		var c domain.CoverageReport
		if err := rows.Scan(&c.ID, &c.TenantID, &c.RunID, &c.Service, &c.LinePct, &c.BranchPct, &c.APIPct, &c.WorkflowPct, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type PolicyRepo struct{ DB *sql.DB }

func (r *PolicyRepo) Save(ctx context.Context, p domain.GatePolicy) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_gate_policies (id, tenant_id, key, kind, thresholds, required, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, key) DO UPDATE SET
			id=EXCLUDED.id, kind=EXCLUDED.kind, thresholds=EXCLUDED.thresholds, required=EXCLUDED.required`,
		p.ID, p.TenantID, p.Key, string(p.Kind), FloatMap(p.Thresholds), p.Required, p.CreatedAt.UTC())
	return err
}
func (r *PolicyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GatePolicy, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, key, kind, thresholds, required, created_at
		FROM qa_gate_policies WHERE tenant_id=$1 ORDER BY key ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GatePolicy{}
	for rows.Next() {
		var p domain.GatePolicy
		var kind string
		var thr FloatMap
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Key, &kind, &thr, &p.Required, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Kind = domain.GateKind(kind)
		p.Thresholds = map[string]float64(thr)
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *PolicyRepo) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.GatePolicy, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, kind, thresholds, required, created_at
		FROM qa_gate_policies WHERE tenant_id=$1 AND key=$2`, tenantID, key)
	var p domain.GatePolicy
	var kind string
	var thr FloatMap
	err := row.Scan(&p.ID, &p.TenantID, &p.Key, &kind, &thr, &p.Required, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.GatePolicy{}, domain.ErrNotFound
		}
		return domain.GatePolicy{}, err
	}
	p.Kind = domain.GateKind(kind)
	p.Thresholds = map[string]float64(thr)
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

type EvalRepo struct{ DB *sql.DB }

func (r *EvalRepo) Save(ctx context.Context, e domain.GateEvaluation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_gate_evals (id, tenant_id, policy_key, run_id, passed, score, details, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.PolicyKey, e.RunID, e.Passed, e.Score, JSONMap(e.Details), e.CreatedAt.UTC())
	return err
}
func (r *EvalRepo) ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.GateEvaluation, error) {
	return r.list(ctx, `SELECT id, tenant_id, policy_key, run_id, passed, score, details, created_at
		FROM qa_gate_evals WHERE tenant_id=$1 AND run_id=$2 ORDER BY created_at DESC`, tenantID, runID)
}
func (r *EvalRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.GateEvaluation, error) {
	return r.list(ctx, `SELECT id, tenant_id, policy_key, run_id, passed, score, details, created_at
		FROM qa_gate_evals WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
}
func (r *EvalRepo) list(ctx context.Context, q string, args ...any) ([]domain.GateEvaluation, error) {
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GateEvaluation{}
	for rows.Next() {
		var e domain.GateEvaluation
		var details JSONMap
		if err := rows.Scan(&e.ID, &e.TenantID, &e.PolicyKey, &e.RunID, &e.Passed, &e.Score, &details, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Details = map[string]any(details)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type CertRepo struct{ DB *sql.DB }

func (r *CertRepo) Save(ctx context.Context, c domain.Certification) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_certifications (
			id, tenant_id, kind, version, commit_sha, status, gate_result_ids, issued_at, expires_at, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, notes=EXCLUDED.notes, expires_at=EXCLUDED.expires_at`,
		c.ID, c.TenantID, string(c.Kind), c.Version, c.CommitSHA, c.Status, UUIDArray(c.GateResults),
		c.IssuedAt.UTC(), nullTime(c.ExpiresAt), c.Notes)
	return err
}
func (r *CertRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Certification, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, kind, version, commit_sha, status, gate_result_ids, issued_at, expires_at, notes
		FROM qa_certifications WHERE tenant_id=$1 ORDER BY issued_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Certification{}
	for rows.Next() {
		c, err := scanCert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *CertRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Certification, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, version, commit_sha, status, gate_result_ids, issued_at, expires_at, notes
		FROM qa_certifications WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	c, err := scanCert(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Certification{}, domain.ErrNotFound
		}
		return domain.Certification{}, err
	}
	return c, nil
}
func scanCert(row scannable) (domain.Certification, error) {
	var c domain.Certification
	var kind string
	var gates UUIDArray
	var expires sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &kind, &c.Version, &c.CommitSHA, &c.Status, &gates, &c.IssuedAt, &expires, &c.Notes)
	if err != nil {
		return domain.Certification{}, err
	}
	c.Kind = domain.CertKind(kind)
	c.GateResults = []uuid.UUID(gates)
	c.ExpiresAt = scanNullTime(expires)
	c.IssuedAt = c.IssuedAt.UTC()
	return c, nil
}

type FlakyRepo struct{ DB *sql.DB }

func (r *FlakyRepo) Save(ctx context.Context, f domain.FlakyRecord) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_flaky (id, tenant_id, test_name, suite_key, fail_count, pass_count, last_status, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, suite_key, test_name) DO UPDATE SET
			id=EXCLUDED.id, fail_count=EXCLUDED.fail_count, pass_count=EXCLUDED.pass_count,
			last_status=EXCLUDED.last_status, updated_at=EXCLUDED.updated_at`,
		f.ID, f.TenantID, f.TestName, f.SuiteKey, f.FailCount, f.PassCount, f.LastStatus, f.UpdatedAt.UTC())
	return err
}
func (r *FlakyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.FlakyRecord, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, test_name, suite_key, fail_count, pass_count, last_status, updated_at
		FROM qa_flaky WHERE tenant_id=$1 ORDER BY updated_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FlakyRecord{}
	for rows.Next() {
		var f domain.FlakyRecord
		if err := rows.Scan(&f.ID, &f.TenantID, &f.TestName, &f.SuiteKey, &f.FailCount, &f.PassCount, &f.LastStatus, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.UpdatedAt = f.UpdatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}
func (r *FlakyRepo) GetByName(ctx context.Context, tenantID uuid.UUID, suiteKey, name string) (domain.FlakyRecord, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, test_name, suite_key, fail_count, pass_count, last_status, updated_at
		FROM qa_flaky WHERE tenant_id=$1 AND suite_key=$2 AND test_name=$3`, tenantID, suiteKey, name)
	var f domain.FlakyRecord
	err := row.Scan(&f.ID, &f.TenantID, &f.TestName, &f.SuiteKey, &f.FailCount, &f.PassCount, &f.LastStatus, &f.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.FlakyRecord{}, domain.ErrNotFound
		}
		return domain.FlakyRecord{}, err
	}
	f.UpdatedAt = f.UpdatedAt.UTC()
	return f, nil
}

type PerfRepo struct{ DB *sql.DB }

func (r *PerfRepo) Save(ctx context.Context, p domain.PerfMetric) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_perf (id, tenant_id, run_id, scenario, p50_ms, p95_ms, p99_ms, error_rate, rps, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.TenantID, p.RunID, p.Scenario, p.P50Ms, p.P95Ms, p.P99Ms, p.ErrorRate, p.RPS, p.CreatedAt.UTC())
	return err
}
func (r *PerfRepo) ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.PerfMetric, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, run_id, scenario, p50_ms, p95_ms, p99_ms, error_rate, rps, created_at
		FROM qa_perf WHERE tenant_id=$1 AND run_id=$2 ORDER BY created_at DESC`, tenantID, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PerfMetric{}
	for rows.Next() {
		var p domain.PerfMetric
		if err := rows.Scan(&p.ID, &p.TenantID, &p.RunID, &p.Scenario, &p.P50Ms, &p.P95Ms, &p.P99Ms, &p.ErrorRate, &p.RPS, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type SecurityRepo struct{ DB *sql.DB }

func (r *SecurityRepo) Save(ctx context.Context, f domain.SecurityFinding) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qa_security_findings (id, tenant_id, run_id, tool, severity, title, target, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		f.ID, f.TenantID, f.RunID, f.Tool, f.Severity, f.Title, f.Target, f.CreatedAt.UTC())
	return err
}
func (r *SecurityRepo) ListByRun(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.SecurityFinding, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, run_id, tool, severity, title, target, created_at
		FROM qa_security_findings WHERE tenant_id=$1 AND run_id=$2 ORDER BY created_at DESC`, tenantID, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SecurityFinding{}
	for rows.Next() {
		var f domain.SecurityFinding
		if err := rows.Scan(&f.ID, &f.TenantID, &f.RunID, &f.Tool, &f.Severity, &f.Title, &f.Target, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.CreatedAt = f.CreatedAt.UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

var (
	_ ports.SuiteRepo    = (*SuiteRepo)(nil)
	_ ports.RunRepo      = (*RunRepo)(nil)
	_ ports.ResultRepo   = (*ResultRepo)(nil)
	_ ports.CoverageRepo = (*CoverageRepo)(nil)
	_ ports.PolicyRepo   = (*PolicyRepo)(nil)
	_ ports.EvalRepo     = (*EvalRepo)(nil)
	_ ports.CertRepo     = (*CertRepo)(nil)
	_ ports.FlakyRepo    = (*FlakyRepo)(nil)
	_ ports.PerfRepo     = (*PerfRepo)(nil)
	_ ports.SecurityRepo = (*SecurityRepo)(nil)
)
