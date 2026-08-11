package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/domain"
)

func (d *Deps) UpsertSuite(ctx context.Context, s domain.Suite) (domain.Suite, error) {
	if err := domain.ValidateSuite(s); err != nil {
		return domain.Suite{}, err
	}
	if s.ID == uuid.Nil {
		if existing, err := d.Suites.GetByKey(ctx, s.TenantID, s.Key); err == nil {
			s.ID = existing.ID
			s.CreatedAt = existing.CreatedAt
		} else {
			s.ID = d.newID()
			s.CreatedAt = d.now()
		}
	}
	s.Enabled = true
	if err := d.Suites.Save(ctx, s); err != nil {
		return domain.Suite{}, err
	}
	return s, nil
}

func (d *Deps) SeedDefaultSuites(ctx context.Context, tenantID uuid.UUID) error {
	defaults := []domain.Suite{
		{Key: "unit-go", Name: "Go unit tests", Kind: domain.SuiteUnit, Path: "services/*/...", Owner: "platform"},
		{Key: "integration-cert", Name: "Launch integration cert", Kind: domain.SuiteIntegration, Path: "tools/integration-cert", Owner: "qa"},
		{Key: "api-smoke", Name: "API smoke", Kind: domain.SuiteAPI, Path: "qa/suites/api", Owner: "qa"},
		{Key: "e2e-customer", Name: "Customer journey", Kind: domain.SuiteE2E, Path: "qa/suites/e2e/customer", Owner: "qa"},
		{Key: "ui-admin-playwright", Name: "Admin Playwright", Kind: domain.SuiteUI, Path: "qa/playwright", Owner: "qa"},
		{Key: "perf-k6-checkout", Name: "Checkout load", Kind: domain.SuitePerf, Path: "qa/k6/checkout_load.js", Owner: "sre"},
		{Key: "security-zap", Name: "OWASP ZAP baseline", Kind: domain.SuiteSecurity, Path: "qa/zap", Owner: "sec"},
		{Key: "a11y-wcag", Name: "WCAG checks", Kind: domain.SuiteA11y, Path: "qa/suites/a11y", Owner: "qa"},
		{Key: "chaos-kafka", Name: "Kafka chaos", Kind: domain.SuiteChaos, Path: "qa/chaos/kafka.yaml", Owner: "sre"},
		{Key: "db-migrations", Name: "Migration checks", Kind: domain.SuiteDB, Path: "qa/suites/db", Owner: "platform"},
		{Key: "infra-helm", Name: "Helm/Terraform validate", Kind: domain.SuiteInfra, Path: "qa/suites/infra", Owner: "platform"},
		{Key: "ai-safety", Name: "AI safety/latency", Kind: domain.SuiteAI, Path: "qa/suites/ai", Owner: "ml"},
		{Key: "obs-slo", Name: "SLO verification", Kind: domain.SuiteObs, Path: "qa/suites/obs", Owner: "sre"},
	}
	for _, s := range defaults {
		s.TenantID = tenantID
		if _, err := d.UpsertSuite(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) SeedDefaultPolicies(ctx context.Context, tenantID uuid.UUID) error {
	pols := []domain.GatePolicy{
		{Key: "unit_coverage", Kind: domain.GateUnitCov, Thresholds: map[string]float64{"minLinePct": 70}, Required: true},
		{Key: "integration", Kind: domain.GateIntegration, Thresholds: map[string]float64{"maxFailed": 0}, Required: true},
		{Key: "performance", Kind: domain.GatePerf, Thresholds: map[string]float64{"maxP95Ms": 500, "maxErrorRate": 0.01}, Required: true},
		{Key: "security", Kind: domain.GateSecurity, Thresholds: map[string]float64{"maxCritical": 0, "maxHigh": 0}, Required: true},
		{Key: "accessibility", Kind: domain.GateA11y, Thresholds: map[string]float64{"maxViolations": 0}, Required: false},
	}
	for _, p := range pols {
		p.ID = d.newID()
		p.TenantID = tenantID
		p.CreatedAt = d.now()
		if err := d.Policies.Save(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) StartRun(ctx context.Context, run domain.TestRun) (domain.TestRun, error) {
	if run.TenantID == uuid.Nil || strings.TrimSpace(run.SuiteKey) == "" {
		return domain.TestRun{}, domain.ErrInvalidArgument
	}
	suite, err := d.Suites.GetByKey(ctx, run.TenantID, run.SuiteKey)
	if err != nil {
		return domain.TestRun{}, err
	}
	run.ID = d.newID()
	run.Kind = suite.Kind
	run.Status = domain.RunRunning
	run.StartedAt = d.now()
	if run.Trigger == "" {
		run.Trigger = "manual"
	}
	if run.Environment == "" {
		run.Environment = "ci"
	}
	if err := d.Runs.Save(ctx, run); err != nil {
		return domain.TestRun{}, err
	}
	d.emit(ctx, run.TenantID, run.ID, domain.EventTestStarted, map[string]any{
		"suiteKey": run.SuiteKey, "kind": string(run.Kind),
	})
	if d.Runner != nil {
		_ = d.Runner.Dispatch(ctx, suite, run)
	}
	return run, nil
}

func (d *Deps) CompleteRun(ctx context.Context, tenantID, runID uuid.UUID, summary domain.RunSummary, cases []domain.TestCaseResult) (domain.TestRun, error) {
	run, err := d.Runs.Get(ctx, tenantID, runID)
	if err != nil {
		return domain.TestRun{}, err
	}
	if err := domain.CanComplete(run); err != nil {
		return domain.TestRun{}, err
	}
	now := d.now()
	run.FinishedAt = &now
	run.Summary = summary
	if summary.Failed > 0 {
		run.Status = domain.RunFailed
	} else {
		run.Status = domain.RunPassed
	}
	if err := d.Runs.Save(ctx, run); err != nil {
		return domain.TestRun{}, err
	}
	for _, c := range cases {
		c.ID = d.newID()
		c.TenantID = tenantID
		c.RunID = runID
		_ = d.Results.Save(ctx, c)
		if c.Status == "flaky" || c.Status == "failed" {
			_ = d.trackFlaky(ctx, tenantID, run.SuiteKey, c)
		}
	}
	d.emit(ctx, tenantID, run.ID, domain.EventTestCompleted, map[string]any{
		"suiteKey": run.SuiteKey, "status": string(run.Status), "failed": summary.Failed,
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "quality.run.completed", map[string]string{"suite": run.SuiteKey, "status": string(run.Status)}, 1)
	}
	return run, nil
}

func (d *Deps) trackFlaky(ctx context.Context, tenantID uuid.UUID, suiteKey string, c domain.TestCaseResult) error {
	rec, err := d.Flaky.GetByName(ctx, tenantID, suiteKey, c.Name)
	if err != nil {
		rec = domain.FlakyRecord{ID: d.newID(), TenantID: tenantID, TestName: c.Name, SuiteKey: suiteKey}
	}
	if c.Status == "failed" {
		rec.FailCount++
	} else {
		rec.PassCount++
	}
	rec.LastStatus = c.Status
	rec.UpdatedAt = d.now()
	return d.Flaky.Save(ctx, rec)
}

func (d *Deps) IngestCoverage(ctx context.Context, c domain.CoverageReport) (domain.CoverageReport, error) {
	if c.TenantID == uuid.Nil || c.Service == "" {
		return domain.CoverageReport{}, domain.ErrInvalidArgument
	}
	c.ID = d.newID()
	c.CreatedAt = d.now()
	if err := d.Coverage.Save(ctx, c); err != nil {
		return domain.CoverageReport{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCoverageGenerated, map[string]any{
		"service": c.Service, "linePct": c.LinePct,
	})
	return c, nil
}

func (d *Deps) IngestPerf(ctx context.Context, p domain.PerfMetric) (domain.PerfMetric, error) {
	if p.TenantID == uuid.Nil || p.RunID == uuid.Nil {
		return domain.PerfMetric{}, domain.ErrInvalidArgument
	}
	p.ID = d.newID()
	p.CreatedAt = d.now()
	if err := d.Perf.Save(ctx, p); err != nil {
		return domain.PerfMetric{}, err
	}
	return p, nil
}

func (d *Deps) IngestSecurityFinding(ctx context.Context, f domain.SecurityFinding) (domain.SecurityFinding, error) {
	if f.TenantID == uuid.Nil || f.Title == "" {
		return domain.SecurityFinding{}, domain.ErrInvalidArgument
	}
	f.ID = d.newID()
	f.CreatedAt = d.now()
	if err := d.Security.Save(ctx, f); err != nil {
		return domain.SecurityFinding{}, err
	}
	return f, nil
}

func (d *Deps) EvaluateGates(ctx context.Context, tenantID, runID uuid.UUID) ([]domain.GateEvaluation, error) {
	run, err := d.Runs.Get(ctx, tenantID, runID)
	if err != nil {
		return nil, err
	}
	pols, err := d.Policies.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	covs, _ := d.Coverage.List(ctx, tenantID)
	perfs, _ := d.Perf.ListByRun(ctx, tenantID, runID)
	findings, _ := d.Security.ListByRun(ctx, tenantID, runID)

	var out []domain.GateEvaluation
	for _, p := range pols {
		ev := domain.GateEvaluation{
			ID: d.newID(), TenantID: tenantID, PolicyKey: p.Key, RunID: runID,
			Details: map[string]any{}, CreatedAt: d.now(),
		}
		switch p.Kind {
		case domain.GateUnitCov:
			line := 0.0
			for _, c := range covs {
				if c.RunID == runID || c.LinePct > line {
					line = c.LinePct
				}
			}
			min := p.Thresholds["minLinePct"]
			ev.Passed, ev.Score = domain.EvaluateUnitCoverage(line, min)
			ev.Details["linePct"] = line
		case domain.GateIntegration:
			ev.Passed = run.Status == domain.RunPassed || (run.Summary.Failed == 0 && run.Status != domain.RunFailed)
			if run.Kind != domain.SuiteIntegration && run.Kind != domain.SuiteAPI && run.Kind != domain.SuiteE2E {
				// non-integration runs: pass if no failures when evaluating batch
				ev.Passed = run.Summary.Failed == 0
			}
			ev.Score = 1
			if !ev.Passed {
				ev.Score = 0
			}
			ev.Details["failed"] = run.Summary.Failed
		case domain.GatePerf:
			p95, errRate := 0.0, 0.0
			if len(perfs) > 0 {
				p95, errRate = perfs[0].P95Ms, perfs[0].ErrorRate
			} else if run.Kind != domain.SuitePerf {
				ev.Passed, ev.Score = true, 1
				ev.Details["skipped"] = true
				out = append(out, ev)
				continue
			}
			ev.Passed, ev.Score = domain.EvaluatePerf(p95, p.Thresholds["maxP95Ms"], errRate, p.Thresholds["maxErrorRate"])
			ev.Details["p95Ms"] = p95
			ev.Details["errorRate"] = errRate
		case domain.GateSecurity:
			crit, high := 0, 0
			for _, f := range findings {
				switch strings.ToLower(f.Severity) {
				case "critical":
					crit++
				case "high":
					high++
				}
			}
			if run.Kind != domain.SuiteSecurity && len(findings) == 0 {
				ev.Passed, ev.Score = true, 1
				ev.Details["skipped"] = true
				out = append(out, ev)
				continue
			}
			ev.Passed, ev.Score = domain.EvaluateSecurity(crit, high, p.Thresholds["maxHigh"] > 0)
			ev.Details["critical"] = crit
			ev.Details["high"] = high
		case domain.GateA11y:
			ev.Passed, ev.Score = true, 1
			ev.Details["note"] = "ingest a11y violations via results"
		default:
			ev.Passed, ev.Score = run.Summary.Failed == 0, 1
		}
		_ = d.Evals.Save(ctx, ev)
		out = append(out, ev)
		if ev.Passed {
			d.emit(ctx, tenantID, ev.ID, domain.EventQualityGatePassed, map[string]any{"policyKey": p.Key, "runId": runID.String()})
		} else {
			d.emit(ctx, tenantID, ev.ID, domain.EventQualityGateFailed, map[string]any{"policyKey": p.Key, "runId": runID.String()})
		}
	}
	return out, nil
}

func (d *Deps) IssueCertification(ctx context.Context, tenantID uuid.UUID, kind domain.CertKind, version, commitSHA, notes string, runIDs []uuid.UUID) (domain.Certification, error) {
	if tenantID == uuid.Nil || version == "" || commitSHA == "" {
		return domain.Certification{}, domain.ErrInvalidArgument
	}
	pols, _ := d.Policies.List(ctx, tenantID)
	var allEvals []domain.GateEvaluation
	var gateIDs []uuid.UUID
	for _, rid := range runIDs {
		evals, err := d.EvaluateGates(ctx, tenantID, rid)
		if err != nil {
			return domain.Certification{}, err
		}
		allEvals = append(allEvals, evals...)
		for _, e := range evals {
			gateIDs = append(gateIDs, e.ID)
		}
	}
	if len(runIDs) > 0 && !domain.AllRequiredGatesPassed(allEvals, pols) {
		return domain.Certification{}, domain.ErrGateFailed
	}
	cert := domain.Certification{
		ID: d.newID(), TenantID: tenantID, Kind: kind, Version: version, CommitSHA: commitSHA,
		Status: "issued", GateResults: gateIDs, IssuedAt: d.now(), Notes: notes,
	}
	if err := d.Certs.Save(ctx, cert); err != nil {
		return domain.Certification{}, err
	}
	d.emit(ctx, tenantID, cert.ID, domain.EventCertificationIssued, map[string]any{
		"kind": string(kind), "version": version, "commitSha": commitSHA,
	})
	d.emit(ctx, tenantID, cert.ID, domain.EventRegressionCompleted, map[string]any{"version": version})
	return cert, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	runs, _ := d.Runs.List(ctx, tenantID)
	certs, _ := d.Certs.List(ctx, tenantID)
	flaky, _ := d.Flaky.List(ctx, tenantID)
	suites, _ := d.Suites.List(ctx, tenantID)
	passed, failed := 0, 0
	for _, r := range runs {
		switch r.Status {
		case domain.RunPassed:
			passed++
		case domain.RunFailed:
			failed++
		}
	}
	return map[string]any{
		"suites": len(suites), "runs": len(runs), "passed": passed, "failed": failed,
		"certificates": len(certs), "flakyTests": len(flaky),
	}, nil
}

func (d *Deps) ListRuns(ctx context.Context, tenantID uuid.UUID) ([]domain.TestRun, error) {
	return d.Runs.List(ctx, tenantID)
}

func (d *Deps) ListCerts(ctx context.Context, tenantID uuid.UUID) ([]domain.Certification, error) {
	return d.Certs.List(ctx, tenantID)
}

func (d *Deps) ListFlaky(ctx context.Context, tenantID uuid.UUID) ([]domain.FlakyRecord, error) {
	return d.Flaky.List(ctx, tenantID)
}
