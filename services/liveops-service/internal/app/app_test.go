package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/app"
	"github.com/nexora/liveops-service/internal/app/memory"
	"github.com/nexora/liveops-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Flags: r.Flags, Configs: r.Configs, Experiments: r.Experiments,
		Events: r.Events, Changes: r.Changes, Rollbacks: r.Rollbacks,
		Outbox: r.Outbox, Cache: r.Cache, Metrics: r.Metrics, AI: r.AI,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestLiveOpsFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	f, err := d.UpsertFlag(ctx, domain.FeatureFlag{
		TenantID: tid, Key: "recs_v2", Enabled: true, Percentage: 100,
		Rules: []domain.TargetRule{{Kind: "country", Values: []string{"TR"}}},
	})
	if err != nil || f.Version != 1 {
		t.Fatal(err)
	}
	evs, err := d.EvaluateFlags(ctx, tid, []string{"recs_v2"}, domain.EvalContext{SubjectID: "u1", Country: "TR"})
	if err != nil || len(evs) != 1 || !evs[0].Enabled {
		t.Fatalf("%+v", evs)
	}

	cfg, err := d.PublishConfig(ctx, domain.ConfigDocument{
		TenantID: tid, Key: "home_layout", Namespace: "home",
		Payload: map[string]any{"banners": 3, "rails": []string{"for_you"}},
	})
	if err != nil || cfg.Version != 1 {
		t.Fatal(err)
	}

	exp, err := d.UpsertExperiment(ctx, domain.Experiment{
		TenantID: tid, Key: "checkout_btn", Kind: "ab", Name: "Checkout CTA",
		Variants: []domain.Variant{{Key: "control", Weight: 50}, {Key: "green", Weight: 50}},
	})
	if err != nil {
		t.Fatal(err)
	}
	exp, err = d.StartExperiment(ctx, tid, exp.Key)
	if err != nil || exp.Status != "running" {
		t.Fatal(err)
	}
	a1, err := d.AssignExperiment(ctx, tid, exp.Key, "subject-9")
	if err != nil {
		t.Fatal(err)
	}
	a2, err := d.AssignExperiment(ctx, tid, exp.Key, "subject-9")
	if err != nil || a1.VariantKey != a2.VariantKey {
		t.Fatal("sticky assign")
	}
	exp, err = d.CompleteExperiment(ctx, tid, exp.Key, map[string]float64{"control": 0.1, "green": 0.15}, true)
	if err != nil || exp.Winner == "" {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	_, err = d.UpsertLiveEvent(ctx, domain.LiveOpsEvent{
		TenantID: tid, Key: "weekend_flash", Kind: "weekend", Title: "Weekend",
		StartsAt: now.Add(-time.Minute), EndsAt: now.Add(2 * time.Hour),
		ConfigPatch: map[string]any{"flash_sale": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = d.ActivateEvent(ctx, tid, "weekend_flash")
	resolved, err := d.ResolveConfig(ctx, tid, "home", domain.EvalContext{SubjectID: "u1"})
	if err != nil || resolved["home_layout"] == nil {
		t.Fatal(err)
	}
	if resolved["flash_sale"] != true {
		t.Fatalf("patch missing %#v", resolved)
	}

	ch, err := d.RequestChange(ctx, domain.ChangeRequest{
		TenantID: tid, Kind: "flag", SubjectKey: "recs_v2", Payload: map[string]any{"pct": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	ch, err = d.DecideChange(ctx, tid, ch.ID, true)
	if err != nil || ch.Status != "approved" {
		t.Fatal(err)
	}

	rb, err := d.Rollback(ctx, tid, "flag", "recs_v2", "latency regression")
	if err != nil || rb.Kind != "flag" {
		t.Fatal(err)
	}
	evs, _ = d.EvaluateFlags(ctx, tid, []string{"recs_v2"}, domain.EvalContext{SubjectID: "u1", Country: "TR"})
	if evs[0].Enabled {
		t.Fatal("should be emergency off")
	}
}
