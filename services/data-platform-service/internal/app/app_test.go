package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/app"
	"github.com/nexora/data-platform-service/internal/app/memory"
	"github.com/nexora/data-platform-service/internal/domain"
)

func testDeps() *app.Deps {
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	return &app.Deps{
		Schemas: repos.Schemas, Events: repos.Events, Streams: repos.Streams,
		Lake: repos.Lake, Warehouse: repos.Warehouse, Realtime: repos.Realtime,
		Experiments: repos.Experiments, Reports: repos.Reports, Obs: repos.Obs,
		Alerts: repos.Alerts, Catalog: repos.Catalog, Quality: repos.Quality,
		Outbox: repos.Outbox, OLAP: repos.OLAP,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestIngestStreamMartsExperiment(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	_, err := d.RegisterSchema(context.Background(), domain.EventSchema{
		TenantID: tenant, Name: "order.placed", Family: domain.FamilyOrder,
		JSONSchema: map[string]any{"required": []any{"orderId", "amountMinor"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.UpsertStreamJob(context.Background(), domain.StreamJob{
		TenantID: tenant, Name: "orders_1m", EventName: "order.placed", WindowSec: 60, Agg: "count",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.UpsertAlertRule(context.Background(), domain.AlertRule{
		TenantID: tenant, Name: "spike", MetricKey: "live.orders", Op: "gte", Threshold: 1, Severity: "warning",
	})
	if err != nil {
		t.Fatal(err)
	}

	e, err := d.IngestEvent(context.Background(), domain.AnalyticsEvent{
		TenantID: tenant, Name: "order.placed", Family: domain.FamilyOrder,
		IdempotencyKey: "idem-1",
		Payload:        map[string]any{"orderId": uuid.New().String(), "amountMinor": float64(4990)},
	})
	if err != nil || !e.Valid {
		t.Fatalf("%v %+v", err, e)
	}
	e2, err := d.IngestEvent(context.Background(), domain.AnalyticsEvent{
		TenantID: tenant, Name: "order.placed", Family: domain.FamilyOrder,
		IdempotencyKey: "idem-1",
		Payload:        map[string]any{"orderId": "x", "amountMinor": float64(1)},
	})
	if err != nil || e2.ID != e.ID {
		t.Fatal("idempotency")
	}

	kpis, err := d.RefreshMarts(context.Background(), tenant)
	if err != nil || len(kpis) == 0 {
		t.Fatalf("%v %v", err, kpis)
	}
	rt, err := d.Realtime.Get(context.Background(), tenant, "live.orders")
	if err != nil || rt.Value < 1 {
		t.Fatalf("%v %+v", err, rt)
	}
	alerts, _ := d.Alerts.ListEvents(context.Background(), tenant, 10)
	if len(alerts) == 0 {
		t.Fatal("expected alert")
	}

	_, err = d.UpsertExperiment(context.Background(), domain.Experiment{
		TenantID: tenant, Key: "home_rank", Name: "Home ranking",
		Variants: []domain.ExperimentVariant{{Name: "control", Weight: 50}, {Name: "treatment", Weight: 50}},
		Status:   "running",
	})
	if err != nil {
		t.Fatal(err)
	}
	a, err := d.AssignExperiment(context.Background(), tenant, "home_rank", uuid.New())
	if err != nil || a.Variant == "" {
		t.Fatalf("%v %+v", err, a)
	}
	exp, err := d.DecideExperiment(context.Background(), tenant, "home_rank", map[string]float64{"control": 0.1, "treatment": 0.2})
	if err != nil || exp.Winner != "treatment" {
		t.Fatalf("%v %+v", err, exp)
	}

	rep, err := d.UpsertReportDef(context.Background(), domain.ReportDef{
		TenantID: tenant, Name: "Executive daily", Kind: "executive", Format: "csv",
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err := d.RunReport(context.Background(), tenant, rep.ID)
	if err != nil || run.Status != "completed" {
		t.Fatalf("%v %+v", err, run)
	}

	_, err = d.IngestObs(context.Background(), domain.ObservabilitySignal{
		TenantID: tenant, Kind: "metric", Service: "order-service", Name: "latency_ms", Value: 42,
	})
	if err != nil {
		t.Fatal(err)
	}
	checks, err := d.RunQualityChecks(context.Background(), tenant, "order.placed")
	if err != nil || len(checks) < 2 {
		t.Fatalf("%v %v", err, checks)
	}
}

func TestSchemaIncompatible(t *testing.T) {
	d := testDeps()
	tenant := uuid.New()
	_, err := d.RegisterSchema(context.Background(), domain.EventSchema{
		TenantID: tenant, Name: "payment.captured", Family: domain.FamilyPayment,
		JSONSchema: map[string]any{"required": []any{"paymentId"}}, Compatibility: domain.CompatBackward,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.RegisterSchema(context.Background(), domain.EventSchema{
		TenantID: tenant, Name: "payment.captured", Family: domain.FamilyPayment,
		JSONSchema: map[string]any{"required": []any{"paymentId", "newRequired"}}, Compatibility: domain.CompatBackward,
	})
	if err != domain.ErrSchemaIncompatible {
		t.Fatalf("want incompatible got %v", err)
	}
}
