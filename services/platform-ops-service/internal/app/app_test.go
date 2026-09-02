package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/app"
	"github.com/nexora/platform-ops-service/internal/app/memory"
	"github.com/nexora/platform-ops-service/internal/domain"
)

func TestOpsFlows(t *testing.T) {
	ctx := context.Background()
	s := memory.NewStore()
	r := memory.NewRepos(s)
	d := &app.Deps{
		Deployments: r.Deployments, Scaling: r.Scaling, Backups: r.Backups,
		Recoveries: r.Recoveries, Alerts: r.Alerts, Costs: r.Costs, SLOs: r.SLOs,
		Outbox: r.Outbox, GitOps: r.GitOps, Cluster: r.Cluster, BackupTool: r.BackupTool,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{}, Registry: memory.NewRegistry(),
	}
	tid := uuid.New()
	dep, err := d.StartDeployment(ctx, domain.Deployment{
		TenantID: tid, Service: "order-service", Environment: "staging",
		Strategy: "canary", ImageTag: "abc123",
	})
	if err != nil || dep.Status != "started" {
		t.Fatal(err)
	}
	dep, err = d.CompleteDeployment(ctx, tid, dep.ID, true)
	if err != nil || dep.Status != "succeeded" {
		t.Fatal(err)
	}
	dep2, err := d.StartDeployment(ctx, domain.Deployment{
		TenantID: tid, Service: "order-service", Environment: "staging", ImageTag: "bad", Strategy: "rolling",
	})
	if err != nil {
		t.Fatal(err)
	}
	dep2, err = d.Rollback(ctx, tid, dep2.ID)
	if err != nil || dep2.Status != "rolled_back" {
		t.Fatal(err)
	}
	_, err = d.Scale(ctx, tid, "order-service", "staging", 3, 8, "lag")
	if err != nil {
		t.Fatal(err)
	}
	b, err := d.RunBackup(ctx, domain.BackupJob{TenantID: tid, Kind: "postgres", Target: "primary"})
	if err != nil || b.Status != "completed" {
		t.Fatal(err)
	}
	rec, err := d.StartRecovery(ctx, domain.RecoveryJob{TenantID: tid, Kind: "region_failover"})
	if err != nil {
		t.Fatal(err)
	}
	rec, err = d.CompleteRecovery(ctx, tid, rec.ID, "dns flipped")
	if err != nil || rec.Status != "completed" {
		t.Fatal(err)
	}
	_, err = d.FireAlert(ctx, domain.AlertEvent{TenantID: tid, Name: "HighErrorRate", Severity: "critical"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.RecordCost(ctx, domain.CostSnapshot{TenantID: tid, Environment: "prod", AmountMinor: 123456, Period: "2026-08"})
	if err != nil {
		t.Fatal(err)
	}
	slo, err := d.RecordSLO(ctx, domain.SLOReport{TenantID: tid, Service: "checkout", Objective: 99.9, Actual: 99.95})
	if err != nil || slo.BudgetLeft <= 0 {
		t.Fatalf("%+v %v", slo, err)
	}
}

func TestTenantCompanyRegistry(t *testing.T) {
	ctx := context.Background()
	d := &app.Deps{Clock: app.SystemClock{}, IDs: app.UUIDGen{}, Registry: memory.NewRegistry()}
	co, err := d.CreateCompany(ctx, "Acme A.S.", "Acme", "TR", "TRY", "sa@nexora")
	if err != nil {
		t.Fatal(err)
	}
	ten, err := d.CreateTenant(ctx, "Acme QC", "acme-qc", co.ID, "shared", "eu-west-1", "sa@nexora")
	if err != nil {
		t.Fatal(err)
	}
	list, err := d.ListTenants(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if list["total"] != 1 {
		t.Fatalf("%+v", list)
	}
	detail, err := d.PatchTenantIsolation(ctx, ten.ID, "hybrid", "sa@nexora")
	if err != nil || detail["isolationMode"] != "hybrid" {
		t.Fatalf("%+v %v", detail, err)
	}
	prop, err := d.ProposeTenantAction(ctx, ten.ID, "tenant_suspend", "billing", "sa1", "sa@nexora")
	if err != nil {
		t.Fatal(err)
	}
	prop, err = d.ResolveTenantProposal(ctx, prop.ID, "approved", "sa@nexora")
	if err != nil || prop.Status != "executed" {
		t.Fatalf("%+v %v", prop, err)
	}
	got, err := d.GetTenant(ctx, ten.ID)
	if err != nil || got["status"] != "suspended" {
		t.Fatalf("%+v %v", got, err)
	}
	roles := d.RolesSnapshot(ctx)
	rs, _ := roles["roles"].([]map[string]any)
	if len(rs) < 10 {
		t.Fatalf("roles=%d", len(rs))
	}
	audit, err := d.AuditSnapshot(ctx, "tenants")
	if err != nil || audit["total"].(int) == 0 {
		t.Fatalf("%+v %v", audit, err)
	}
}
