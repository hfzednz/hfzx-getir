package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/app"
	"github.com/nexora/superapp-service/internal/app/memory"
	"github.com/nexora/superapp-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Modules: r.Modules, Manifests: r.Manifests, Installs: r.Installs,
		Permissions: r.Permissions, Listings: r.Listings, Ratings: r.Ratings,
		Widgets: r.Widgets, Monetization: r.Monetization, Launches: r.Launches,
		Outbox: r.Outbox, LiveOps: r.LiveOps, AI: r.AI, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{}, ShellVersion: "1.0.0",
	}
}

func TestSuperAppFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()
	subject := "user-opaque-1"

	if err := d.SeedMiniApps(ctx, tid); err != nil {
		t.Fatal(err)
	}
	mods, _ := d.Modules.List(ctx, tid)
	if len(mods) < 17 {
		t.Fatalf("expected mini apps, got %d", len(mods))
	}

	inst, err := d.InstallModule(ctx, tid, subject, "qc")
	if err != nil || inst.Status != domain.InstallActive {
		t.Fatalf("%+v %v", inst, err)
	}
	_, err = d.InstallModule(ctx, tid, subject, "food")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.GrantPermission(ctx, tid, subject, inst.ModuleID, domain.PermPayments)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.AddWidget(ctx, domain.WidgetPlacement{
		TenantID: tid, SubjectID: subject, ModuleID: inst.ModuleID, Slot: "home", Order: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	shell, err := d.ResolveShell(ctx, tid, subject, "1.0.0")
	if err != nil || len(shell.Modules) < 2 {
		t.Fatalf("%+v %v", shell, err)
	}

	ev, err := d.LaunchMiniApp(ctx, tid, subject, "qc")
	if err != nil || ev.ID == uuid.Nil {
		t.Fatal(err)
	}

	store, err := d.BrowseStore(ctx, tid, "commerce")
	if err != nil || len(store) < 1 {
		t.Fatal(store, err)
	}

	_, err = d.RateModule(ctx, domain.StoreRating{
		TenantID: tid, ModuleID: inst.ModuleID, SubjectID: subject, Score: 5, Comment: "fast",
	})
	if err != nil {
		t.Fatal(err)
	}

	keys, _ := d.Recommend(ctx, tid, subject, 3)
	if len(keys) == 0 {
		t.Fatal("expected recommendations")
	}

	found, _ := d.SearchModules(ctx, tid, "food")
	if len(found) == 0 {
		t.Fatal("search")
	}

	_, err = d.UpsertMonetization(ctx, domain.MonetizationRule{
		TenantID: tid, ModuleID: inst.ModuleID, CommissionBps: 1500, PartnerShareBps: 8500,
	})
	if err != nil {
		t.Fatal(err)
	}

	// publish new version then update + rollback
	mod, _ := d.Modules.GetByKey(ctx, tid, "qc")
	_, err = d.PublishManifest(ctx, domain.ModuleManifest{
		TenantID: tid, ModuleID: mod.ID, Version: "1.1.0",
		EntryPoint: "flutter://mini/qc", MinShellVersion: "1.0.0",
		Permissions: []string{domain.PermNavigation}, Signature: "sig-qc-2", Checksum: "sha-qc-2",
		BundleURI: "https://cdn.nexora.example/mini/qc/1.1.0.apk",
	})
	if err != nil {
		t.Fatal(err)
	}
	up, err := d.UpdateModule(ctx, tid, subject, "qc")
	if err != nil || up.Version != "1.1.0" {
		t.Fatalf("%+v %v", up, err)
	}
	rb, err := d.RollbackInstall(ctx, tid, subject, "qc")
	if err != nil || rb.Version != "1.0.0" {
		t.Fatalf("%+v %v", rb, err)
	}

	_, err = d.RemoveModule(ctx, tid, subject, "food")
	if err != nil {
		t.Fatal(err)
	}

	st, _ := d.AdminStats(ctx, tid)
	if st["miniApps"].(int) < 17 {
		t.Fatal(st)
	}
}
