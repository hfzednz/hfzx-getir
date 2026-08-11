package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/enterprise-ops-service/internal/domain"
)

func (d *Deps) UpsertOrg(ctx context.Context, n domain.OrgNode) (domain.OrgNode, error) {
	if err := domain.ValidateOrg(n); err != nil {
		return domain.OrgNode{}, err
	}
	now := d.now()
	if n.ID == uuid.Nil {
		if existing, err := d.Org.GetByCode(ctx, n.TenantID, n.Code); err == nil {
			n.ID = existing.ID
			n.CreatedAt = existing.CreatedAt
		} else {
			n.ID = d.newID()
			n.CreatedAt = now
		}
	}
	n.Active = true
	n.UpdatedAt = now
	if err := d.Org.Save(ctx, n); err != nil {
		return domain.OrgNode{}, err
	}
	return n, nil
}

func (d *Deps) UpsertPolicy(ctx context.Context, p domain.Policy) (domain.Policy, error) {
	if err := domain.ValidatePolicy(p); err != nil {
		return domain.Policy{}, err
	}
	now := d.now()
	if existing, err := d.Policies.GetByKey(ctx, p.TenantID, p.Key); err == nil {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
		if p.Status == "" {
			p.Status = existing.Status
		}
		if p.Version == "" {
			p.Version = existing.Version
		}
	} else {
		p.ID = d.newID()
		p.CreatedAt = now
		if p.Status == "" {
			p.Status = domain.PolicyDraft
		}
		if p.Version == "" {
			p.Version = "1.0.0"
		}
	}
	p.UpdatedAt = now
	if err := d.Policies.Save(ctx, p); err != nil {
		return domain.Policy{}, err
	}
	return p, nil
}

func (d *Deps) ApprovePolicy(ctx context.Context, tenantID uuid.UUID, key, approver string) (domain.Policy, error) {
	p, err := d.Policies.GetByKey(ctx, tenantID, key)
	if err != nil {
		return domain.Policy{}, err
	}
	if p.Status != domain.PolicyDraft && p.Status != domain.PolicyInReview {
		return domain.Policy{}, domain.ErrIllegalTransition
	}
	if d.Security != nil {
		if ok, err := d.Security.PolicyChangeAllowed(ctx, tenantID, key); err == nil && !ok {
			return domain.Policy{}, domain.ErrForbidden
		}
	}
	if approver == "" {
		return domain.Policy{}, domain.ErrApprovalRequired
	}
	now := d.now()
	p.Status = domain.PolicyApproved
	p.ApprovedBy = approver
	p.ApprovedAt = &now
	p.UpdatedAt = now
	if err := d.Policies.Save(ctx, p); err != nil {
		return domain.Policy{}, err
	}
	d.emit(ctx, tenantID, p.ID, domain.EventPolicyApproved, map[string]any{
		"key": p.Key, "version": p.Version, "approver": approver,
	})
	return p, nil
}

func (d *Deps) CreatePortfolio(ctx context.Context, p domain.Portfolio) (domain.Portfolio, error) {
	if p.TenantID == uuid.Nil || p.Code == "" || p.Name == "" {
		return domain.Portfolio{}, domain.ErrInvalidArgument
	}
	p.ID = d.newID()
	p.CreatedAt = d.now()
	if err := d.Portfolios.Save(ctx, p); err != nil {
		return domain.Portfolio{}, err
	}
	return p, nil
}

func (d *Deps) CreateProgram(ctx context.Context, p domain.Program) (domain.Program, error) {
	if p.TenantID == uuid.Nil || p.PortfolioID == uuid.Nil || p.Code == "" || p.Name == "" {
		return domain.Program{}, domain.ErrInvalidArgument
	}
	if _, err := d.Portfolios.Get(ctx, p.TenantID, p.PortfolioID); err != nil {
		return domain.Program{}, err
	}
	p.ID = d.newID()
	p.CreatedAt = d.now()
	if err := d.Programs.Save(ctx, p); err != nil {
		return domain.Program{}, err
	}
	return p, nil
}

func (d *Deps) CreateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	if p.TenantID == uuid.Nil || p.ProgramID == uuid.Nil || p.Code == "" || p.Name == "" {
		return domain.Project{}, domain.ErrInvalidArgument
	}
	if _, err := d.Programs.Get(ctx, p.TenantID, p.ProgramID); err != nil {
		return domain.Project{}, err
	}
	now := d.now()
	p.ID = d.newID()
	if p.Status == "" {
		p.Status = domain.ProjectPlanned
	}
	if p.Currency == "" {
		p.Currency = "TRY"
	}
	p.Health = domain.ProjectHealth(0, 0)
	p.CreatedAt = now
	p.UpdatedAt = now
	if err := d.Projects.Save(ctx, p); err != nil {
		return domain.Project{}, err
	}
	d.emit(ctx, p.TenantID, p.ID, domain.EventProjectCreated, map[string]any{
		"code": p.Code, "programId": p.ProgramID.String(),
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "enterprise.project.created", map[string]string{"code": p.Code}, 1)
	}
	return p, nil
}

func (d *Deps) AddMilestone(ctx context.Context, m domain.Milestone) (domain.Milestone, error) {
	if m.TenantID == uuid.Nil || m.ProjectID == uuid.Nil || m.Name == "" {
		return domain.Milestone{}, domain.ErrInvalidArgument
	}
	if _, err := d.Projects.Get(ctx, m.TenantID, m.ProjectID); err != nil {
		return domain.Milestone{}, err
	}
	m.ID = d.newID()
	m.CreatedAt = d.now()
	if err := d.Milestones.Save(ctx, m); err != nil {
		return domain.Milestone{}, err
	}
	return m, nil
}

func (d *Deps) CreateObjective(ctx context.Context, o domain.Objective) (domain.Objective, error) {
	if o.TenantID == uuid.Nil || o.Period == "" || o.Title == "" {
		return domain.Objective{}, domain.ErrInvalidArgument
	}
	o.ID = d.newID()
	o.CreatedAt = d.now()
	if err := d.Objectives.Save(ctx, o); err != nil {
		return domain.Objective{}, err
	}
	return o, nil
}

func (d *Deps) UpsertKeyResult(ctx context.Context, kr domain.KeyResult) (domain.KeyResult, error) {
	if kr.TenantID == uuid.Nil || kr.ObjectiveID == uuid.Nil || kr.Title == "" {
		return domain.KeyResult{}, domain.ErrInvalidArgument
	}
	kr.ID = d.newID()
	kr.UpdatedAt = d.now()
	if err := d.KeyResults.Save(ctx, kr); err != nil {
		return domain.KeyResult{}, err
	}
	return kr, nil
}

func (d *Deps) UpsertKPI(ctx context.Context, k domain.KPI) (domain.KPI, error) {
	if k.TenantID == uuid.Nil || k.Key == "" || k.Name == "" || k.Period == "" {
		return domain.KPI{}, domain.ErrInvalidArgument
	}
	k.ID = d.newID()
	k.UpdatedAt = d.now()
	if err := d.KPIs.Save(ctx, k); err != nil {
		return domain.KPI{}, err
	}
	return k, nil
}

func (d *Deps) IdentifyRisk(ctx context.Context, r domain.Risk) (domain.Risk, error) {
	if r.TenantID == uuid.Nil || r.Code == "" || r.Title == "" || r.Category == "" {
		return domain.Risk{}, domain.ErrInvalidArgument
	}
	score, err := domain.RiskScore(r.Likelihood, r.Impact)
	if err != nil {
		return domain.Risk{}, err
	}
	now := d.now()
	r.ID = d.newID()
	r.Score = score
	r.Status = "open"
	r.CreatedAt = now
	r.UpdatedAt = now
	if err := d.Risks.Save(ctx, r); err != nil {
		return domain.Risk{}, err
	}
	d.emit(ctx, r.TenantID, r.ID, domain.EventRiskIdentified, map[string]any{
		"code": r.Code, "category": string(r.Category), "score": r.Score,
	})
	return r, nil
}

func (d *Deps) UpsertContinuity(ctx context.Context, p domain.ContinuityPlan) (domain.ContinuityPlan, error) {
	if p.TenantID == uuid.Nil || p.Key == "" || p.Name == "" {
		return domain.ContinuityPlan{}, domain.ErrInvalidArgument
	}
	now := d.now()
	if existing, err := d.Continuity.GetByKey(ctx, p.TenantID, p.Key); err == nil {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
		p.Status = existing.Status
		p.ActivatedAt = existing.ActivatedAt
	} else {
		p.ID = d.newID()
		p.CreatedAt = now
		if p.Status == "" {
			p.Status = "draft"
		}
	}
	p.UpdatedAt = now
	if err := d.Continuity.Save(ctx, p); err != nil {
		return domain.ContinuityPlan{}, err
	}
	return p, nil
}

func (d *Deps) ActivateContinuity(ctx context.Context, tenantID uuid.UUID, key string) (domain.ContinuityPlan, error) {
	p, err := d.Continuity.GetByKey(ctx, tenantID, key)
	if err != nil {
		return domain.ContinuityPlan{}, err
	}
	if p.Status != "approved" && p.Status != "draft" {
		// allow draft→active for emergency in sandbox; prefer approved
	}
	now := d.now()
	p.Status = "active"
	p.ActivatedAt = &now
	p.UpdatedAt = now
	if err := d.Continuity.Save(ctx, p); err != nil {
		return domain.ContinuityPlan{}, err
	}
	d.emit(ctx, tenantID, p.ID, domain.EventContinuityPlanActivated, map[string]any{
		"key": p.Key, "priority": p.Priority, "criticalService": p.CriticalSvc,
	})
	return p, nil
}

func (d *Deps) CreateAudit(ctx context.Context, a domain.AuditEngagement) (domain.AuditEngagement, error) {
	if a.TenantID == uuid.Nil || a.Code == "" || a.Title == "" {
		return domain.AuditEngagement{}, domain.ErrInvalidArgument
	}
	a.ID = d.newID()
	if a.Kind == "" {
		a.Kind = "internal"
	}
	a.Status = "planned"
	a.CreatedAt = d.now()
	if a.ScheduledAt.IsZero() {
		a.ScheduledAt = d.now()
	}
	if err := d.Audits.Save(ctx, a); err != nil {
		return domain.AuditEngagement{}, err
	}
	return a, nil
}

func (d *Deps) CompleteAudit(ctx context.Context, tenantID, auditID uuid.UUID) (domain.AuditEngagement, error) {
	a, err := d.Audits.Get(ctx, tenantID, auditID)
	if err != nil {
		return domain.AuditEngagement{}, err
	}
	if a.Status == "completed" {
		return domain.AuditEngagement{}, domain.ErrIllegalTransition
	}
	now := d.now()
	a.Status = "completed"
	a.CompletedAt = &now
	if err := d.Audits.Save(ctx, a); err != nil {
		return domain.AuditEngagement{}, err
	}
	d.emit(ctx, tenantID, a.ID, domain.EventAuditCompleted, map[string]any{
		"code": a.Code, "kind": a.Kind,
	})
	return a, nil
}

func (d *Deps) AddFinding(ctx context.Context, f domain.AuditFinding) (domain.AuditFinding, error) {
	if f.TenantID == uuid.Nil || f.AuditID == uuid.Nil || f.Title == "" {
		return domain.AuditFinding{}, domain.ErrInvalidArgument
	}
	if _, err := d.Audits.Get(ctx, f.TenantID, f.AuditID); err != nil {
		return domain.AuditFinding{}, err
	}
	f.ID = d.newID()
	if f.Status == "" {
		f.Status = "open"
	}
	if f.Severity == "" {
		f.Severity = "medium"
	}
	f.CreatedAt = d.now()
	if err := d.Findings.Save(ctx, f); err != nil {
		return domain.AuditFinding{}, err
	}
	return f, nil
}

func (d *Deps) ScheduleMeeting(ctx context.Context, m domain.Meeting) (domain.Meeting, error) {
	if m.TenantID == uuid.Nil || m.Title == "" || m.Kind == "" || m.StartsAt.IsZero() {
		return domain.Meeting{}, domain.ErrInvalidArgument
	}
	m.ID = d.newID()
	m.Status = "scheduled"
	m.CreatedAt = d.now()
	if err := d.Meetings.Save(ctx, m); err != nil {
		return domain.Meeting{}, err
	}
	d.emit(ctx, m.TenantID, m.ID, domain.EventMeetingScheduled, map[string]any{
		"kind": string(m.Kind), "title": m.Title,
	})
	return m, nil
}

func (d *Deps) RecordDecision(ctx context.Context, dec domain.Decision) (domain.Decision, error) {
	if dec.TenantID == uuid.Nil || dec.Title == "" || dec.DecidedBy == "" {
		return domain.Decision{}, domain.ErrInvalidArgument
	}
	dec.ID = d.newID()
	dec.CreatedAt = d.now()
	if err := d.Decisions.Save(ctx, dec); err != nil {
		return domain.Decision{}, err
	}
	d.emit(ctx, dec.TenantID, dec.ID, domain.EventDecisionRecorded, map[string]any{
		"title": dec.Title, "decidedBy": dec.DecidedBy,
	})
	return dec, nil
}

func (d *Deps) UpsertKnowledge(ctx context.Context, doc domain.KnowledgeDoc) (domain.KnowledgeDoc, error) {
	if doc.TenantID == uuid.Nil || doc.Key == "" || doc.Title == "" || doc.URI == "" {
		return domain.KnowledgeDoc{}, domain.ErrInvalidArgument
	}
	doc.ID = d.newID()
	if doc.Kind == "" {
		doc.Kind = "wiki"
	}
	doc.CreatedAt = d.now()
	if err := d.Knowledge.Save(ctx, doc); err != nil {
		return domain.KnowledgeDoc{}, err
	}
	return doc, nil
}

func (d *Deps) UpsertResource(ctx context.Context, r domain.ResourcePlan) (domain.ResourcePlan, error) {
	if r.TenantID == uuid.Nil || r.TeamCode == "" || r.Period == "" {
		return domain.ResourcePlan{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	if r.CapacityFTE > 0 {
		r.Utilization = r.AllocatedFTE / r.CapacityFTE
	}
	r.UpdatedAt = d.now()
	if err := d.Resources.Save(ctx, r); err != nil {
		return domain.ResourcePlan{}, err
	}
	return r, nil
}

func (d *Deps) ExecutiveDashboard(ctx context.Context, tenantID uuid.UUID, role string) (map[string]any, error) {
	projects, _ := d.Projects.List(ctx, tenantID)
	risks, _ := d.Risks.List(ctx, tenantID)
	policies, _ := d.Policies.List(ctx, tenantID)
	audits, _ := d.Audits.List(ctx, tenantID)
	kpis, _ := d.KPIs.List(ctx, tenantID)
	plans, _ := d.Continuity.List(ctx, tenantID)
	openRisks, redProjects, approvedPolicies, activeBCP := 0, 0, 0, 0
	for _, r := range risks {
		if r.Status == "open" {
			openRisks++
		}
	}
	for _, p := range projects {
		if p.Health == "red" {
			redProjects++
		}
	}
	for _, p := range policies {
		if p.Status == domain.PolicyApproved {
			approvedPolicies++
		}
	}
	for _, p := range plans {
		if p.Status == "active" {
			activeBCP++
		}
	}
	dash := map[string]any{
		"role": role, "projects": len(projects), "redProjects": redProjects,
		"openRisks": openRisks, "approvedPolicies": approvedPolicies,
		"audits": len(audits), "kpis": len(kpis), "activeContinuityPlans": activeBCP,
	}
	if d.AI != nil {
		if preds, err := d.AI.RiskPrediction(ctx, tenantID); err == nil {
			dash["aiRiskHints"] = preds
		}
	}
	return dash, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	org, _ := d.Org.List(ctx, tenantID)
	policies, _ := d.Policies.List(ctx, tenantID)
	projects, _ := d.Projects.List(ctx, tenantID)
	risks, _ := d.Risks.List(ctx, tenantID)
	audits, _ := d.Audits.List(ctx, tenantID)
	meetings, _ := d.Meetings.List(ctx, tenantID)
	return map[string]any{
		"orgNodes": len(org), "policies": len(policies), "projects": len(projects),
		"risks": len(risks), "audits": len(audits), "meetings": len(meetings),
	}, nil
}
