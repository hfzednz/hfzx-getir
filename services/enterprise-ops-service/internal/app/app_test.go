package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/app"
	"github.com/nexora/enterprise-ops-service/internal/app/memory"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Org: r.OrgR, Policies: r.PolicyR, Portfolios: r.PortfolioR,
		Programs: r.ProgramR, Projects: r.ProjectR, Milestones: r.MilestoneR,
		Objectives: r.ObjectiveR, KeyResults: r.KeyResultR, KPIs: r.KPIR,
		Risks: r.RiskR, Continuity: r.ContinuityR, Audits: r.AuditR,
		Findings: r.FindingR, Meetings: r.MeetingR, Decisions: r.DecisionR,
		Knowledge: r.KnowledgeR, Resources: r.ResourceR, Outbox: r.OutboxR,
		Security: r.Security, AI: r.AI, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestEnterpriseOpsFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	co, err := d.UpsertOrg(ctx, domain.OrgNode{
		TenantID: tid, Kind: domain.OrgCompany, Code: "NEXORA", Name: "Nexora Holding", CountryCode: "TR",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.UpsertOrg(ctx, domain.OrgNode{
		TenantID: tid, Kind: domain.OrgBusinessUnit, Code: "QC", Name: "Quick Commerce",
		ParentID: &co.ID, CountryCode: "TR",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.UpsertPolicy(ctx, domain.Policy{
		TenantID: tid, Key: "code-of-conduct", Title: "Code of Conduct", Kind: domain.PolicyCorporate,
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := d.ApprovePolicy(ctx, tid, "code-of-conduct", "board-chair")
	if err != nil || p.Status != domain.PolicyApproved {
		t.Fatalf("%+v %v", p, err)
	}

	pf, err := d.CreatePortfolio(ctx, domain.Portfolio{TenantID: tid, Code: "STRAT", Name: "Strategic"})
	if err != nil {
		t.Fatal(err)
	}
	prog, err := d.CreateProgram(ctx, domain.Program{TenantID: tid, PortfolioID: pf.ID, Code: "DIG", Name: "Digital"})
	if err != nil {
		t.Fatal(err)
	}
	proj, err := d.CreateProject(ctx, domain.Project{
		TenantID: tid, ProgramID: prog.ID, Code: "P-100", Name: "City Launch", BudgetMinor: 1_000_000_00, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.AddMilestone(ctx, domain.Milestone{
		TenantID: tid, ProjectID: proj.ID, Name: "Go-Live", DueAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	obj, err := d.CreateObjective(ctx, domain.Objective{TenantID: tid, Period: "2026-Q3", Title: "Operational excellence"})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = d.UpsertKeyResult(ctx, domain.KeyResult{TenantID: tid, ObjectiveID: obj.ID, Title: "OTIF", Target: 95, Current: 91, Unit: "%"})
	_, _ = d.UpsertKPI(ctx, domain.KPI{TenantID: tid, Key: "nps", Name: "NPS", Value: 42, Target: 50, Period: "2026-08"})

	risk, err := d.IdentifyRisk(ctx, domain.Risk{
		TenantID: tid, Code: "R-1", Title: "Capacity shortfall", Category: domain.RiskOperational,
		Likelihood: 3, Impact: 4,
	})
	if err != nil || risk.Score != 12 {
		t.Fatalf("%+v %v", risk, err)
	}

	_, err = d.UpsertContinuity(ctx, domain.ContinuityPlan{
		TenantID: tid, Key: "checkout-critical", Name: "Checkout Continuity",
		RTOHours: 1, RPOHours: 0, Priority: 1, CriticalSvc: "checkout", Status: "approved",
	})
	if err != nil {
		t.Fatal(err)
	}
	act, err := d.ActivateContinuity(ctx, tid, "checkout-critical")
	if err != nil || act.Status != "active" {
		t.Fatalf("%+v %v", act, err)
	}

	audit, err := d.CreateAudit(ctx, domain.AuditEngagement{TenantID: tid, Code: "IA-26", Title: "Ops Controls"})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = d.AddFinding(ctx, domain.AuditFinding{TenantID: tid, AuditID: audit.ID, Title: "Segregation gap", Severity: "high", CAPA: "RACI update"})
	done, err := d.CompleteAudit(ctx, tid, audit.ID)
	if err != nil || done.Status != "completed" {
		t.Fatal(err)
	}

	_, err = d.ScheduleMeeting(ctx, domain.Meeting{
		TenantID: tid, Kind: domain.MeetingExecutive, Title: "Ops Review",
		StartsAt: time.Now().UTC().Add(24 * time.Hour), Agenda: "BCP drill",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.RecordDecision(ctx, domain.Decision{
		TenantID: tid, Title: "Approve city expansion", Body: "Proceed IST-2", DecidedBy: "CEO", VotesFor: 5, VotesAgainst: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _ = d.UpsertKnowledge(ctx, domain.KnowledgeDoc{TenantID: tid, Key: "bcp-playbook", Title: "BCP Playbook", Kind: "playbook", URI: "kb://bcp"})
	_, _ = d.UpsertResource(ctx, domain.ResourcePlan{TenantID: tid, TeamCode: "ops", Period: "2026-Q3", CapacityFTE: 10, AllocatedFTE: 8})

	dash, _ := d.ExecutiveDashboard(ctx, tid, "CEO")
	if dash["openRisks"].(int) < 1 || dash["activeContinuityPlans"].(int) < 1 {
		t.Fatal(dash)
	}
	st, _ := d.AdminStats(ctx, tid)
	if st["projects"].(int) < 1 || st["policies"].(int) < 1 {
		t.Fatal(st)
	}
}
