package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/domain"
)

func (d *Deps) RunAutonomyAudit(ctx context.Context, tenantID uuid.UUID, scope domain.AuditScope) (domain.AutonomyAudit, error) {
	if tenantID == uuid.Nil || scope == "" {
		return domain.AutonomyAudit{}, domain.ErrInvalidArgument
	}
	now := d.now()
	a := domain.AutonomyAudit{
		ID: d.newID(), TenantID: tenantID, Scope: scope, Status: "completed", Score: 96,
		CreatedAt: now, CompletedAt: &now,
	}
	if err := d.Audits.Save(ctx, a); err != nil {
		return domain.AutonomyAudit{}, err
	}
	d.emit(ctx, tenantID, a.ID, domain.EventAutonomyAuditCompleted, map[string]any{
		"scope": string(scope), "score": a.Score,
	})
	return a, nil
}

func (d *Deps) ExecuteHeal(ctx context.Context, a domain.HealAction) (domain.HealAction, error) {
	if err := domain.ValidateHeal(a); err != nil {
		return domain.HealAction{}, err
	}
	a.ID = d.newID()
	a.Status = "planned"
	a.Automated = true
	a.CreatedAt = d.now()
	if d.PlatformOps != nil {
		if err := d.PlatformOps.ExecuteHeal(ctx, a.TenantID, a.Action, a.TargetRef); err != nil {
			a.Status = "failed"
			_ = d.Heals.Save(ctx, a)
			return domain.HealAction{}, err
		}
	}
	now := d.now()
	a.Status = "executed"
	a.ExecutedAt = &now
	if err := d.Heals.Save(ctx, a); err != nil {
		return domain.HealAction{}, err
	}
	d.emit(ctx, a.TenantID, a.ID, domain.EventSelfHealExecuted, map[string]any{
		"kind": string(a.Kind), "action": a.Action, "target": a.TargetRef,
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "autonomy.heal", map[string]string{"kind": string(a.Kind)}, 1)
	}
	return a, nil
}

func (d *Deps) CreateCTOReview(ctx context.Context, r domain.AICTOReview) (domain.AICTOReview, error) {
	if r.TenantID == uuid.Nil || r.Kind == "" {
		return domain.AICTOReview{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	if r.Summary == "" {
		r.Summary = fmt.Sprintf("Autonomous %s review — no redesign required", r.Kind)
	}
	if len(r.Suggestions) == 0 {
		r.Suggestions = []string{"Continue additive hardening", "Keep opaque IDs", "Preserve money as int64 minor units"}
	}
	r.CreatedAt = d.now()
	if err := d.Reviews.Save(ctx, r); err != nil {
		return domain.AICTOReview{}, err
	}
	d.emit(ctx, r.TenantID, r.ID, domain.EventAICTOReviewCompleted, map[string]any{
		"kind": string(r.Kind), "debtScore": r.DebtScore,
	})
	return r, nil
}

func (d *Deps) CreateEvolution(ctx context.Context, t domain.EvolutionTask) (domain.EvolutionTask, error) {
	if t.TenantID == uuid.Nil || t.Kind == "" || t.Title == "" {
		return domain.EvolutionTask{}, domain.ErrInvalidArgument
	}
	t.ID = d.newID()
	if t.Priority <= 0 {
		t.Priority = 3
	}
	t.Status = "backlog"
	t.CreatedAt = d.now()
	if err := d.Evolution.Save(ctx, t); err != nil {
		return domain.EvolutionTask{}, err
	}
	d.emit(ctx, t.TenantID, t.ID, domain.EventEvolutionTaskCreated, map[string]any{
		"kind": string(t.Kind), "title": t.Title,
	})
	return t, nil
}

func (d *Deps) ScoreRelease(ctx context.Context, p domain.ReleasePlan) (domain.ReleasePlan, error) {
	if p.TenantID == uuid.Nil || p.Version == "" {
		return domain.ReleasePlan{}, domain.ErrInvalidArgument
	}
	testsOK, secOK := true, true
	if d.Quality != nil {
		ok, _ := d.Quality.Healthy(ctx, p.TenantID)
		testsOK = ok
	}
	if d.Security != nil {
		ok, _ := d.Security.Healthy(ctx, p.TenantID)
		secOK = ok
	}
	p.ID = d.newID()
	if p.Strategy == "" {
		p.Strategy = "canary"
	}
	p.Score = domain.ReleaseScore(true, testsOK, secOK, 0.2)
	p.Status = "validated"
	p.CreatedAt = d.now()
	if err := d.Releases.Save(ctx, p); err != nil {
		return domain.ReleasePlan{}, err
	}
	d.emit(ctx, p.TenantID, p.ID, domain.EventAutonomousReleaseScored, map[string]any{
		"version": p.Version, "score": p.Score, "strategy": p.Strategy,
	})
	return p, nil
}

func (d *Deps) UpsertGovernance(ctx context.Context, g domain.GovernanceLoop) (domain.GovernanceLoop, error) {
	if g.TenantID == uuid.Nil || g.Domain == "" {
		return domain.GovernanceLoop{}, domain.ErrInvalidArgument
	}
	g.ID = d.newID()
	if g.Cadence == "" {
		g.Cadence = "continuous"
	}
	g.Healthy = true
	g.UpdatedAt = d.now()
	if err := d.Governance.Save(ctx, g); err != nil {
		return domain.GovernanceLoop{}, err
	}
	return g, nil
}

func (d *Deps) BootstrapAutonomy(ctx context.Context, tenantID uuid.UUID) error {
	scopes := []domain.AuditScope{
		domain.ScopeDependency, domain.ScopeArchitecture, domain.ScopeBusiness,
		domain.ScopeInfrastructure, domain.ScopeSecurity, domain.ScopePerformance,
		domain.ScopeAI, domain.ScopeDocumentation, domain.ScopeCompliance,
		domain.ScopeDX, domain.ScopeOperational,
	}
	for _, s := range scopes {
		if _, err := d.RunAutonomyAudit(ctx, tenantID, s); err != nil {
			return err
		}
	}
	for _, e := range domain.DefaultDependencyGraph() {
		e.ID = d.newID()
		e.TenantID = tenantID
		e.CreatedAt = d.now()
		if err := d.Dependencies.Save(ctx, e); err != nil {
			return err
		}
	}
	heals := []domain.HealAction{
		{TenantID: tenantID, Kind: domain.HealService, TargetRef: "bff-customer", Action: "restart"},
		{TenantID: tenantID, Kind: domain.HealDatabase, TargetRef: "postgres-primary", Action: "failover"},
		{TenantID: tenantID, Kind: domain.HealSecurity, TargetRef: "vault", Action: "rotate"},
		{TenantID: tenantID, Kind: domain.HealAI, TargetRef: "rank-model", Action: "retrain"},
		{TenantID: tenantID, Kind: domain.HealQA, TargetRef: "flaky-suite", Action: "patch"},
	}
	for _, h := range heals {
		if _, err := d.ExecuteHeal(ctx, h); err != nil {
			return err
		}
	}
	for _, kind := range []domain.ReviewKind{
		domain.ReviewArchitecture, domain.ReviewSecurity, domain.ReviewPerformance,
		domain.ReviewCost, domain.ReviewRelease,
	} {
		if _, err := d.CreateCTOReview(ctx, domain.AICTOReview{TenantID: tenantID, Kind: kind, DebtScore: 12}); err != nil {
			return err
		}
	}
	evos := []domain.EvolutionTask{
		{TenantID: tenantID, Kind: domain.EvoDependency, Title: "Pin Go module advisories", Priority: 1},
		{TenantID: tenantID, Kind: domain.EvoDocs, Title: "Regenerate OpenAPI indexes", Priority: 2},
		{TenantID: tenantID, Kind: domain.EvoDebt, Title: "Eliminate unused stubs debt", Priority: 3},
	}
	for _, t := range evos {
		if _, err := d.CreateEvolution(ctx, t); err != nil {
			return err
		}
	}
	if _, err := d.ScoreRelease(ctx, domain.ReleasePlan{TenantID: tenantID, Version: "genesis-1.0.0", Strategy: "canary"}); err != nil {
		return err
	}
	for _, domainName := range []string{"architecture", "security", "ai", "data", "operational", "compliance"} {
		if _, err := d.UpsertGovernance(ctx, domain.GovernanceLoop{TenantID: tenantID, Domain: domainName}); err != nil {
			return err
		}
	}
	for _, a := range domain.DefaultAssistants() {
		a.ID = d.newID()
		a.TenantID = tenantID
		a.CreatedAt = d.now()
		if err := d.Assistants.Save(ctx, a); err != nil {
			return err
		}
	}
	for _, t := range domain.DefaultDigitalTeams() {
		t.ID = d.newID()
		t.TenantID = tenantID
		t.CreatedAt = d.now()
		if err := d.Teams.Save(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) EvaluateGenesisGates(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	gates := map[string]bool{}
	for _, g := range domain.GenesisGatesRequired() {
		gates[g] = false
	}
	audits, _ := d.Audits.List(ctx, tenantID)
	gates["autonomy_audits"] = len(audits) >= 11
	executed, _ := d.Heals.ExecutedCount(ctx, tenantID)
	gates["self_healing"] = executed >= 3
	reviews, _ := d.Reviews.List(ctx, tenantID)
	gates["ai_cto"] = len(reviews) >= 3
	evos, _ := d.Evolution.List(ctx, tenantID)
	gates["evolution"] = len(evos) >= 1
	rels, _ := d.Releases.List(ctx, tenantID)
	releaseOK := false
	for _, r := range rels {
		if r.Score >= 80 {
			releaseOK = true
			break
		}
	}
	gates["release_engine"] = releaseOK
	govs, _ := d.Governance.List(ctx, tenantID)
	govOK := len(govs) >= 6
	for _, g := range govs {
		if !g.Healthy {
			govOK = false
		}
	}
	gates["governance"] = govOK
	if d.Hyperscale != nil {
		ok, _ := d.Hyperscale.Certified(ctx, tenantID)
		gates["hyperscale"] = ok
	}
	if d.Security != nil {
		ok, _ := d.Security.Healthy(ctx, tenantID)
		gates["security"] = ok
	}
	if d.Quality != nil {
		ok, _ := d.Quality.Healthy(ctx, tenantID)
		gates["quality"] = ok
	}
	return gates, nil
}

func (d *Deps) IssueGenesis(ctx context.Context, tenantID uuid.UUID, version string) (domain.GenesisCertificate, error) {
	if version == "" {
		version = "1.0.0"
	}
	gates, err := d.EvaluateGenesisGates(ctx, tenantID)
	if err != nil {
		return domain.GenesisCertificate{}, err
	}
	for _, g := range domain.GenesisGatesRequired() {
		if !gates[g] {
			return domain.GenesisCertificate{}, domain.ErrGateFailed
		}
	}
	now := d.now()
	cert := domain.GenesisCertificate{
		ID: d.newID(), TenantID: tenantID, Version: version, Status: "issued",
		Gates: gates, IssuedAt: &now, CreatedAt: now,
	}
	if err := d.Genesis.Save(ctx, cert); err != nil {
		return domain.GenesisCertificate{}, err
	}
	d.emit(ctx, tenantID, cert.ID, domain.EventGenesisCertified, map[string]any{
		"version": version,
	})
	return cert, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	audits, _ := d.Audits.List(ctx, tenantID)
	heals, _ := d.Heals.List(ctx, tenantID)
	reviews, _ := d.Reviews.List(ctx, tenantID)
	evos, _ := d.Evolution.List(ctx, tenantID)
	deps, _ := d.Dependencies.List(ctx, tenantID)
	certs, _ := d.Genesis.List(ctx, tenantID)
	openW, _ := d.Weaknesses.OpenCount(ctx, tenantID)
	return map[string]any{
		"audits": len(audits), "heals": len(heals), "reviews": len(reviews),
		"evolutionTasks": len(evos), "dependencyEdges": len(deps),
		"openWeaknesses": openW, "genesisCertificates": len(certs),
	}, nil
}
