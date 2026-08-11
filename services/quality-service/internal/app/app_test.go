package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/app"
	"github.com/nexora/quality-service/internal/app/memory"
	"github.com/nexora/quality-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Suites: r.Suites, Runs: r.Runs, Results: r.Results, Coverage: r.Coverage,
		Policies: r.Policies, Evals: r.Evals, Certs: r.Certs, Flaky: r.Flaky,
		Perf: r.Perf, Security: r.Security, Outbox: r.Outbox, Runner: r.Runner, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestQualityCertificationFlow(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.SeedDefaultSuites(ctx, tid); err != nil {
		t.Fatal(err)
	}
	if err := d.SeedDefaultPolicies(ctx, tid); err != nil {
		t.Fatal(err)
	}

	run, err := d.StartRun(ctx, domain.TestRun{
		TenantID: tid, SuiteKey: "integration-cert", Trigger: "ci", CommitSHA: "abc123", Branch: "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err = d.CompleteRun(ctx, tid, run.ID, domain.RunSummary{Total: 10, Passed: 10, DurationMs: 1200}, []domain.TestCaseResult{
		{Name: "registry_ok", Status: "passed"},
		{Name: "identity_ok", Status: "passed"},
	})
	if err != nil || run.Status != domain.RunPassed {
		t.Fatal(err)
	}

	_, err = d.IngestCoverage(ctx, domain.CoverageReport{
		TenantID: tid, RunID: run.ID, Service: "checkout-service", LinePct: 82, BranchPct: 70, APIPct: 90,
	})
	if err != nil {
		t.Fatal(err)
	}

	perfRun, err := d.StartRun(ctx, domain.TestRun{TenantID: tid, SuiteKey: "perf-k6-checkout", Trigger: "ci", CommitSHA: "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.CompleteRun(ctx, tid, perfRun.ID, domain.RunSummary{Total: 1, Passed: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.IngestPerf(ctx, domain.PerfMetric{
		TenantID: tid, RunID: perfRun.ID, Scenario: "checkout_load", P50Ms: 80, P95Ms: 220, P99Ms: 350, ErrorRate: 0.001, RPS: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	secRun, err := d.StartRun(ctx, domain.TestRun{TenantID: tid, SuiteKey: "security-zap", Trigger: "ci", CommitSHA: "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.CompleteRun(ctx, tid, secRun.ID, domain.RunSummary{Total: 1, Passed: 1}, nil)
	if err != nil {
		t.Fatal(err)
	}

	cert, err := d.IssueCertification(ctx, tid, domain.CertProduction, "1.0.0", "abc123", "wave quality", []uuid.UUID{run.ID, perfRun.ID, secRun.ID})
	if err != nil || cert.Status != "issued" {
		t.Fatalf("%+v %v", cert, err)
	}

	st, err := d.AdminStats(ctx, tid)
	if err != nil || st["certificates"].(int) < 1 {
		t.Fatalf("%v", st)
	}
}
