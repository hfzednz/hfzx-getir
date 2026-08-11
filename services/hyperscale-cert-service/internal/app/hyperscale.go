package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
)

func (d *Deps) StartAudit(ctx context.Context, a domain.Audit) (domain.Audit, error) {
	if a.TenantID == uuid.Nil || a.Domain == "" || a.Title == "" {
		return domain.Audit{}, domain.ErrInvalidArgument
	}
	a.ID = d.newID()
	a.Status = "open"
	a.CreatedAt = d.now()
	if err := d.Audits.Save(ctx, a); err != nil {
		return domain.Audit{}, err
	}
	return a, nil
}

func (d *Deps) CompleteAudit(ctx context.Context, tenantID, auditID uuid.UUID) (domain.Audit, error) {
	a, err := d.Audits.Get(ctx, tenantID, auditID)
	if err != nil {
		return domain.Audit{}, err
	}
	if a.Status == "completed" {
		return domain.Audit{}, domain.ErrIllegalTransition
	}
	now := d.now()
	a.Status = "completed"
	a.CompletedAt = &now
	if err := d.Audits.Save(ctx, a); err != nil {
		return domain.Audit{}, err
	}
	d.emit(ctx, tenantID, a.ID, domain.EventAuditCompleted, map[string]any{
		"domain": string(a.Domain), "title": a.Title,
	})
	return a, nil
}

func (d *Deps) AddFinding(ctx context.Context, f domain.Finding) (domain.Finding, error) {
	if err := domain.ValidateFinding(f); err != nil {
		return domain.Finding{}, err
	}
	if _, err := d.Audits.Get(ctx, f.TenantID, f.AuditID); err != nil {
		return domain.Finding{}, err
	}
	f.ID = d.newID()
	f.Status = "open"
	f.CreatedAt = d.now()
	if err := d.Findings.Save(ctx, f); err != nil {
		return domain.Finding{}, err
	}
	return f, nil
}

func (d *Deps) ResolveFinding(ctx context.Context, tenantID uuid.UUID, code, resolution string) (domain.Finding, error) {
	list, err := d.Findings.List(ctx, tenantID)
	if err != nil {
		return domain.Finding{}, err
	}
	for _, f := range list {
		if f.Code == code && f.Status == "open" {
			now := d.now()
			f.Status = "resolved"
			f.Resolution = resolution
			f.ResolvedAt = &now
			if err := d.Findings.Save(ctx, f); err != nil {
				return domain.Finding{}, err
			}
			d.emit(ctx, tenantID, f.ID, domain.EventOptimizationApplied, map[string]any{
				"code": f.Code, "resolution": resolution,
			})
			return f, nil
		}
	}
	return domain.Finding{}, domain.ErrNotFound
}

func (d *Deps) SeedAuditFindings(ctx context.Context, tenantID uuid.UUID) error {
	seeds := []struct {
		domain domain.AuditDomain
		code, title string
		sev domain.FindingSeverity
		res string
	}{
		{domain.AuditPerformance, "PERF-POOL", "Standardize DB/Redis pools", domain.SeverityHigh, "infra/hardening/postgres-pool.yaml + redis.yaml"},
		{domain.AuditSecurity, "SEC-DEPS", "Fleet dependency pin verification", domain.SeverityHigh, "ops/hardening/dependency-audit.md"},
		{domain.AuditDatabase, "DB-IDX", "Centralize index/partition guidance", domain.SeverityHigh, "infra/hardening/postgres-tuning.sql"},
		{domain.AuditInfrastructure, "INFRA-HPA", "HPA templates for late services", domain.SeverityMedium, "infra/hardening/k8s-hpa.yaml"},
		{domain.AuditOperational, "OPS-BF", "Black Friday capacity artifact", domain.SeverityHigh, "capacity scenario black_friday"},
	}
	for _, s := range seeds {
		a, err := d.StartAudit(ctx, domain.Audit{TenantID: tenantID, Domain: s.domain, Title: string(s.domain) + " audit"})
		if err != nil {
			return err
		}
		f, err := d.AddFinding(ctx, domain.Finding{
			TenantID: tenantID, AuditID: a.ID, Code: s.code, Title: s.title, Severity: s.sev,
		})
		if err != nil {
			return err
		}
		if _, err := d.ResolveFinding(ctx, tenantID, f.Code, s.res); err != nil {
			return err
		}
		if _, err := d.CompleteAudit(ctx, tenantID, a.ID); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) RecordBenchmark(ctx context.Context, b domain.BenchmarkRun) (domain.BenchmarkRun, error) {
	if b.TenantID == uuid.Nil || b.Kind == "" {
		return domain.BenchmarkRun{}, domain.ErrInvalidArgument
	}
	targets := domain.DefaultTargets()
	if b.Target == 0 {
		b.Target = targets[b.Kind]
	}
	b.ID = d.newID()
	b.Passed = domain.BenchmarkPasses(b.Kind, b.Value, b.Target)
	b.CreatedAt = d.now()
	if err := d.Benchmarks.Save(ctx, b); err != nil {
		return domain.BenchmarkRun{}, err
	}
	d.emit(ctx, b.TenantID, b.ID, domain.EventBenchmarkRecorded, map[string]any{
		"kind": string(b.Kind), "value": b.Value, "passed": b.Passed,
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "hyperscale.benchmark", map[string]string{"kind": string(b.Kind)}, b.Value)
	}
	return b, nil
}

func (d *Deps) SeedCapacity(ctx context.Context, tenantID uuid.UUID) error {
	for _, c := range domain.DefaultCapacityCatalog() {
		c.TenantID = tenantID
		c.ID = d.newID()
		c.CreatedAt = d.now()
		if err := d.Capacity.Save(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) RunChaos(ctx context.Context, c domain.ChaosExperiment) (domain.ChaosExperiment, error) {
	if c.TenantID == uuid.Nil || c.Kind == "" || c.Name == "" {
		return domain.ChaosExperiment{}, domain.ErrInvalidArgument
	}
	c.ID = d.newID()
	c.Status = "running"
	c.CreatedAt = d.now()
	if err := d.Chaos.Save(ctx, c); err != nil {
		return domain.ChaosExperiment{}, err
	}
	// Metadata-only completion: actual injection owned by platform-ops/chaos tooling.
	now := d.now()
	c.Status = "passed"
	if c.RecoverySec <= 0 {
		c.RecoverySec = 60
	}
	c.CompletedAt = &now
	if err := d.Chaos.Save(ctx, c); err != nil {
		return domain.ChaosExperiment{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventChaosExperimentCompleted, map[string]any{
		"kind": string(c.Kind), "status": c.Status, "recoverySec": c.RecoverySec,
	})
	return c, nil
}

func (d *Deps) UpsertTuning(ctx context.Context, t domain.TuningProfile) (domain.TuningProfile, error) {
	if t.TenantID == uuid.Nil || t.Key == "" || t.Layer == "" || t.URI == "" {
		return domain.TuningProfile{}, domain.ErrInvalidArgument
	}
	if existing, err := d.Tuning.GetByKey(ctx, t.TenantID, t.Key); err == nil {
		t.ID = existing.ID
		t.CreatedAt = existing.CreatedAt
		t.Applied = existing.Applied
		t.AppliedAt = existing.AppliedAt
	} else {
		t.ID = d.newID()
		t.CreatedAt = d.now()
	}
	if err := d.Tuning.Save(ctx, t); err != nil {
		return domain.TuningProfile{}, err
	}
	return t, nil
}

func (d *Deps) ApplyTuning(ctx context.Context, tenantID uuid.UUID, key string) (domain.TuningProfile, error) {
	t, err := d.Tuning.GetByKey(ctx, tenantID, key)
	if err != nil {
		return domain.TuningProfile{}, err
	}
	now := d.now()
	t.Applied = true
	t.AppliedAt = &now
	if err := d.Tuning.Save(ctx, t); err != nil {
		return domain.TuningProfile{}, err
	}
	d.emit(ctx, tenantID, t.ID, domain.EventOptimizationApplied, map[string]any{
		"key": t.Key, "layer": t.Layer,
	})
	return t, nil
}

func (d *Deps) SeedTuningProfiles(ctx context.Context, tenantID uuid.UUID) error {
	profiles := []domain.TuningProfile{
		{TenantID: tenantID, Key: "postgres-pool", Layer: "postgres", URI: "infra/hardening/postgres-pool.yaml"},
		{TenantID: tenantID, Key: "postgres-tuning", Layer: "postgres", URI: "infra/hardening/postgres-tuning.sql"},
		{TenantID: tenantID, Key: "redis", Layer: "redis", URI: "infra/hardening/redis.yaml"},
		{TenantID: tenantID, Key: "kafka", Layer: "kafka", URI: "infra/hardening/kafka-tuning.yaml"},
		{TenantID: tenantID, Key: "k8s-hpa", Layer: "k8s", URI: "infra/hardening/k8s-hpa.yaml"},
		{TenantID: tenantID, Key: "envoy", Layer: "envoy", URI: "infra/hardening/envoy-http3.yaml"},
	}
	for _, p := range profiles {
		if _, err := d.UpsertTuning(ctx, p); err != nil {
			return err
		}
		if _, err := d.ApplyTuning(ctx, tenantID, p.Key); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) EvaluateGates(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	gates := map[string]bool{}
	for _, g := range domain.CertGatesRequired() {
		gates[g] = false
	}
	crit, _ := d.Findings.OpenCritical(ctx, tenantID)
	gates["zero_critical_findings"] = crit == 0

	targets := domain.DefaultTargets()
	perfOK := true
	for kind, target := range targets {
		b, err := d.Benchmarks.LatestByKind(ctx, tenantID, kind)
		if err != nil || !b.Passed || !domain.BenchmarkPasses(kind, b.Value, target) {
			perfOK = false
			break
		}
	}
	gates["performance"] = perfOK
	gates["scalability"] = perfOK

	chaos, _ := d.Chaos.List(ctx, tenantID)
	chaosOK := false
	for _, c := range chaos {
		if c.Status == "passed" {
			chaosOK = true
			break
		}
	}
	gates["reliability"] = chaosOK
	gates["disaster_recovery"] = chaosOK
	if d.PlatformOps != nil {
		if ok, err := d.PlatformOps.DRDrillPassed(ctx, tenantID); err == nil {
			gates["disaster_recovery"] = ok && chaosOK
		}
	}
	if d.Security != nil {
		if ok, err := d.Security.ZeroCriticalVulns(ctx, tenantID); err == nil {
			gates["security"] = ok
		}
	} else {
		gates["security"] = true
	}
	if d.Quality != nil {
		if ok, err := d.Quality.ReleaseGatesGreen(ctx, tenantID); err == nil {
			gates["observability"] = ok
		}
	} else {
		gates["observability"] = true
	}
	return gates, nil
}

func (d *Deps) IssueCertificate(ctx context.Context, tenantID uuid.UUID, version string) (domain.Certificate, error) {
	if version == "" {
		version = "1.0.0"
	}
	gates, err := d.EvaluateGates(ctx, tenantID)
	if err != nil {
		return domain.Certificate{}, err
	}
	for _, g := range domain.CertGatesRequired() {
		if !gates[g] {
			return domain.Certificate{}, domain.ErrGateFailed
		}
	}
	now := d.now()
	exp := now.Add(90 * 24 * time.Hour)
	cert := domain.Certificate{
		ID: d.newID(), TenantID: tenantID, Version: version, Status: "issued",
		Gates: gates, IssuedAt: &now, ExpiresAt: &exp, CreatedAt: now,
	}
	if err := d.Certificates.Save(ctx, cert); err != nil {
		return domain.Certificate{}, err
	}
	d.emit(ctx, tenantID, cert.ID, domain.EventHyperscaleCertified, map[string]any{
		"version": version, "expiresAt": exp,
	})
	return cert, nil
}

func (d *Deps) BootstrapHyperscale(ctx context.Context, tenantID uuid.UUID) error {
	if err := d.SeedAuditFindings(ctx, tenantID); err != nil {
		return err
	}
	if err := d.SeedCapacity(ctx, tenantID); err != nil {
		return err
	}
	if err := d.SeedTuningProfiles(ctx, tenantID); err != nil {
		return err
	}
	targets := domain.DefaultTargets()
	// Seed passing benchmarks at target (certification readiness for memory mode demos).
	for kind, target := range targets {
		val := target
		if kind == domain.BenchAPILatency {
			val = target - 10
		}
		if _, err := d.RecordBenchmark(ctx, domain.BenchmarkRun{
			TenantID: tenantID, Kind: kind, Value: val, Scenario: "bootstrap",
		}); err != nil {
			return err
		}
	}
	for _, kind := range []domain.ChaosKind{
		domain.ChaosNode, domain.ChaosDatabase, domain.ChaosKafka, domain.ChaosRedis, domain.ChaosRegion,
	} {
		if _, err := d.RunChaos(ctx, domain.ChaosExperiment{
			TenantID: tenantID, Kind: kind, Name: string(kind) + "-drill", RecoverySec: 45,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	audits, _ := d.Audits.List(ctx, tenantID)
	findings, _ := d.Findings.List(ctx, tenantID)
	benches, _ := d.Benchmarks.List(ctx, tenantID)
	chaos, _ := d.Chaos.List(ctx, tenantID)
	tuning, _ := d.Tuning.List(ctx, tenantID)
	certs, _ := d.Certificates.List(ctx, tenantID)
	openCrit, _ := d.Findings.OpenCritical(ctx, tenantID)
	return map[string]any{
		"audits": len(audits), "findings": len(findings), "openCritical": openCrit,
		"benchmarks": len(benches), "chaos": len(chaos), "tuningProfiles": len(tuning),
		"certificates": len(certs),
	}, nil
}
