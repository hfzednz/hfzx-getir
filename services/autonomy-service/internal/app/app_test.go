package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app"
	"github.com/nexora/autonomy-service/internal/app/memory"
	"github.com/nexora/autonomy-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Audits: r.Audits, Weaknesses: r.Weaknesses, Heals: r.Heals,
		Reviews: r.Reviews, Evolution: r.Evolution, Releases: r.Releases,
		Governance: r.Governance, Assistants: r.Assistants, Teams: r.Teams,
		Dependencies: r.Dependencies, Genesis: r.Genesis, Outbox: r.Outbox,
		Hyperscale: r.Hyperscale, PlatformOps: r.Platform, Quality: r.Quality,
		Security: r.Security, LiveOps: r.LiveOps, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestGenesisCertification(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.BootstrapAutonomy(ctx, tid); err != nil {
		t.Fatal(err)
	}

	gates, err := d.EvaluateGenesisGates(ctx, tid)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range domain.GenesisGatesRequired() {
		if !gates[g] {
			t.Fatalf("gate %s failed: %+v", g, gates)
		}
	}

	cert, err := d.IssueGenesis(ctx, tid, "1.0.0")
	if err != nil || cert.Status != "issued" {
		t.Fatalf("%+v %v", cert, err)
	}

	st, _ := d.AdminStats(ctx, tid)
	if st["audits"].(int) < 11 || st["genesisCertificates"].(int) < 1 {
		t.Fatal(st)
	}

	// break hyperscale port → re-issue must fail
	if mock, ok := d.Hyperscale.(*memory.MockHyperscale); ok {
		mock.Set(false)
	}
	if _, err := d.IssueGenesis(ctx, tid, "1.0.1"); err != domain.ErrGateFailed {
		t.Fatalf("want gate failed got %v", err)
	}
}
