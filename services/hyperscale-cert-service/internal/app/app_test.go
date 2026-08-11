package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/app"
	"github.com/nexora/hyperscale-cert-service/internal/app/memory"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Audits: r.AuditR, Findings: r.FindingR, Benchmarks: r.BenchR,
		Capacity: r.CapR, Chaos: r.ChaosR, Tuning: r.TuningR,
		Certificates: r.CertR, Outbox: r.OutboxR,
		Quality: r.Quality, PlatformOps: r.Platform, Security: r.Security,
		Metrics: r.Metrics, Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestHyperscaleCertification(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.BootstrapHyperscale(ctx, tid); err != nil {
		t.Fatal(err)
	}

	gates, err := d.EvaluateGates(ctx, tid)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range domain.CertGatesRequired() {
		if !gates[g] {
			t.Fatalf("gate %s failed: %+v", g, gates)
		}
	}

	cert, err := d.IssueCertificate(ctx, tid, "1.0.0")
	if err != nil || cert.Status != "issued" {
		t.Fatalf("%+v %v", cert, err)
	}

	st, _ := d.AdminStats(ctx, tid)
	if st["openCritical"].(int) != 0 || st["certificates"].(int) < 1 {
		t.Fatal(st)
	}

	// failing bench should block re-issue after recording bad latency
	_, err = d.RecordBenchmark(ctx, domain.BenchmarkRun{
		TenantID: tid, Kind: domain.BenchAPILatency, Value: 500, Scenario: "regression",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.IssueCertificate(ctx, tid, "1.0.1"); err != domain.ErrGateFailed {
		t.Fatalf("want gate failed got %v", err)
	}
}
