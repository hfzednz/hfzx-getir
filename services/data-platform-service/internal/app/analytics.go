package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/domain"
)

// UpsertExperiment creates/updates an A/B experiment.
func (d *Deps) UpsertExperiment(ctx context.Context, e domain.Experiment) (domain.Experiment, error) {
	if e.TenantID == uuid.Nil || e.Key == "" || len(e.Variants) == 0 {
		return e, domain.ErrInvalidArgument
	}
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.Status == "" {
		e.Status = "draft"
	}
	if e.PrimaryKPI == "" {
		e.PrimaryKPI = domain.KPIConversion
	}
	e.UpdatedAt = d.now()
	return e, d.Experiments.Save(ctx, e)
}

// AssignExperiment assigns a subject to a variant.
func (d *Deps) AssignExperiment(ctx context.Context, tenantID uuid.UUID, experimentKey string, subjectID uuid.UUID) (domain.ExperimentAssignment, error) {
	var zero domain.ExperimentAssignment
	exp, err := d.Experiments.Get(ctx, tenantID, experimentKey)
	if err != nil {
		return zero, err
	}
	if existing, ok, _ := d.Experiments.GetAssignment(ctx, tenantID, exp.ID, subjectID); ok {
		return existing, nil
	}
	a := domain.ExperimentAssignment{
		TenantID: tenantID, ExperimentID: exp.ID, SubjectID: subjectID,
		Variant: domain.AssignVariant(subjectID, exp.Variants), AssignedAt: d.now(),
	}
	return a, d.Experiments.SaveAssignment(ctx, a)
}

// DecideExperiment picks a winner from variant scores.
func (d *Deps) DecideExperiment(ctx context.Context, tenantID uuid.UUID, experimentKey string, scores map[string]float64) (domain.Experiment, error) {
	exp, err := d.Experiments.Get(ctx, tenantID, experimentKey)
	if err != nil {
		return exp, err
	}
	exp.Winner = domain.DecideWinner(scores)
	exp.Status = "decided"
	now := d.now()
	exp.EndedAt = &now
	exp.UpdatedAt = now
	if err := d.Experiments.Save(ctx, exp); err != nil {
		return exp, err
	}
	d.emit(ctx, tenantID, exp.ID, domain.EventExperimentDecided, map[string]any{
		"key": exp.Key, "winner": exp.Winner, "scores": scores,
	})
	return exp, nil
}

// UpsertReportDef saves a report definition.
func (d *Deps) UpsertReportDef(ctx context.Context, r domain.ReportDef) (domain.ReportDef, error) {
	if r.TenantID == uuid.Nil || r.Name == "" {
		return r, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		r.ID = d.newID()
	}
	if r.Format == "" {
		r.Format = "json"
	}
	r.Active = true
	r.UpdatedAt = d.now()
	return r, d.Reports.SaveDef(ctx, r)
}

// RunReport materializes a report from KPIs/facts.
func (d *Deps) RunReport(ctx context.Context, tenantID, reportID uuid.UUID) (domain.ReportRun, error) {
	var zero domain.ReportRun
	defs, err := d.Reports.ListDefs(ctx, tenantID)
	if err != nil {
		return zero, err
	}
	var def *domain.ReportDef
	for i := range defs {
		if defs[i].ID == reportID {
			def = &defs[i]
			break
		}
	}
	if def == nil {
		return zero, domain.ErrNotFound
	}
	kpis, _ := d.Warehouse.ListKPIs(ctx, tenantID)
	now := d.now()
	run := domain.ReportRun{
		ID: d.newID(), TenantID: tenantID, ReportID: reportID, Status: "completed",
		Location: fmt.Sprintf("s3://nexora-reports/%s/%s.%s", tenantID, reportID, def.Format),
		RowCount: len(kpis), CreatedAt: now, CompletedAt: &now,
	}
	_ = d.Reports.SaveRun(ctx, run)
	d.emit(ctx, tenantID, run.ID, domain.EventReportGenerated, map[string]any{
		"reportId": reportID.String(), "format": def.Format, "rows": run.RowCount,
	})
	return run, nil
}

// IngestObs stores a trace/log/metric signal.
func (d *Deps) IngestObs(ctx context.Context, s domain.ObservabilitySignal) (domain.ObservabilitySignal, error) {
	if s.TenantID == uuid.Nil || s.Kind == "" || s.Service == "" {
		return s, domain.ErrInvalidArgument
	}
	switch s.Kind {
	case "span", "log", "metric":
	default:
		return s, domain.ErrInvalidArgument
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	if s.OccurredAt.IsZero() {
		s.OccurredAt = d.now()
	}
	if err := d.Obs.SaveSignal(ctx, s); err != nil {
		return s, err
	}
	if s.Kind == "metric" {
		_, _ = d.Realtime.Incr(ctx, s.TenantID, "obs."+s.Service+"."+s.Name, s.Value, d.now())
		_ = d.evaluateAlerts(ctx, s.TenantID)
	}
	return s, nil
}

// UpsertAlertRule saves an alert rule.
func (d *Deps) UpsertAlertRule(ctx context.Context, r domain.AlertRule) (domain.AlertRule, error) {
	if r.TenantID == uuid.Nil || r.Name == "" || r.MetricKey == "" {
		return r, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		r.ID = d.newID()
	}
	if r.Severity == "" {
		r.Severity = "warning"
	}
	r.Enabled = true
	r.UpdatedAt = d.now()
	return r, d.Alerts.SaveRule(ctx, r)
}

func (d *Deps) evaluateAlerts(ctx context.Context, tenantID uuid.UUID) error {
	rules, err := d.Alerts.ListRules(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		m, err := d.Realtime.Get(ctx, tenantID, rule.MetricKey)
		if err != nil {
			continue
		}
		if domain.EvalThreshold(m.Value, rule.Op, rule.Threshold) {
			ev := domain.AlertEvent{
				ID: d.newID(), TenantID: tenantID, RuleID: rule.ID, MetricKey: rule.MetricKey,
				Value: m.Value, Severity: rule.Severity,
				Message: fmt.Sprintf("%s: %s %s %v (actual %v)", rule.Name, rule.MetricKey, rule.Op, rule.Threshold, m.Value),
				CreatedAt: d.now(),
			}
			_ = d.Alerts.SaveEvent(ctx, ev)
			d.emit(ctx, tenantID, ev.ID, domain.EventAlertFired, map[string]any{
				"rule": rule.Name, "metric": rule.MetricKey, "value": m.Value,
			})
		}
	}
	return nil
}

// RunQualityChecks evaluates freshness/completeness for an asset.
func (d *Deps) RunQualityChecks(ctx context.Context, tenantID uuid.UUID, assetName string) ([]domain.QualityCheck, error) {
	now := d.now()
	events, _ := d.Events.List(ctx, tenantID, assetName, 100)
	completeness := 1.0
	if len(events) == 0 {
		completeness = 0
	}
	fresh := 1.0
	if len(events) > 0 {
		age := now.Sub(events[0].IngestedAt).Hours()
		if age > 24 {
			fresh = 0.5
		}
		if age > 72 {
			fresh = 0
		}
	}
	checks := []domain.QualityCheck{
		{ID: d.newID(), TenantID: tenantID, AssetName: assetName, CheckType: "completeness", Passed: completeness >= 0.9, Score: completeness, CreatedAt: now},
		{ID: d.newID(), TenantID: tenantID, AssetName: assetName, CheckType: "freshness", Passed: fresh >= 0.9, Score: fresh, CreatedAt: now},
	}
	uniq := 1.0
	hashes := map[string]int{}
	for _, e := range events {
		hashes[e.PayloadHash]++
	}
	dups := 0
	for _, n := range hashes {
		if n > 1 {
			dups++
		}
	}
	if len(events) > 0 {
		uniq = 1 - float64(dups)/float64(len(events))
	}
	checks = append(checks, domain.QualityCheck{
		ID: d.newID(), TenantID: tenantID, AssetName: assetName, CheckType: "uniqueness",
		Passed: uniq >= 0.95, Score: uniq, CreatedAt: now,
	})
	for _, c := range checks {
		_ = d.Quality.Save(ctx, c)
		if !c.Passed {
			d.emit(ctx, tenantID, c.ID, domain.EventQualityFailed, map[string]any{
				"asset": assetName, "check": c.CheckType, "score": c.Score,
			})
		}
	}
	return checks, nil
}

// UpsertCatalogAsset registers a catalog entry.
func (d *Deps) UpsertCatalogAsset(ctx context.Context, a domain.CatalogAsset) (domain.CatalogAsset, error) {
	if a.TenantID == uuid.Nil || a.Name == "" || a.Type == "" {
		return a, domain.ErrInvalidArgument
	}
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	if a.Classification == "" {
		a.Classification = "internal"
	}
	a.UpdatedAt = d.now()
	return a, d.Catalog.SaveAsset(ctx, a)
}

// AdminStats returns dashboard counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	events, _ := d.Events.List(ctx, tenantID, "", 1)
	_ = events
	schemas, _ := d.Schemas.List(ctx, tenantID, "")
	jobs, _ := d.Streams.ListJobs(ctx, tenantID)
	kpis, _ := d.Warehouse.ListKPIs(ctx, tenantID)
	alerts, _ := d.Alerts.ListEvents(ctx, tenantID, 50)
	rt, _ := d.Realtime.List(ctx, tenantID)
	return map[string]any{
		"schemas": len(schemas), "streamJobs": len(jobs), "kpis": len(kpis),
		"recentAlerts": len(alerts), "realtimeKeys": len(rt), "tenantId": tenantID.String(),
	}, nil
}

// GDPRDelete soft-deletes by marking quality + emitting (payload wipe on listed events for user).
func (d *Deps) GDPRDeleteUser(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	list, err := d.Events.List(ctx, tenantID, "", 5000)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range list {
		if e.UserID != nil && *e.UserID == userID {
			e.Payload = map[string]any{"redacted": true}
			e.Valid = false
			e.Error = "gdpr_erasure"
			_ = d.Events.Save(ctx, e)
			n++
		}
	}
	return n, nil
}
