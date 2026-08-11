package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/domain"
)

func (d *Deps) StartDeployment(ctx context.Context, dep domain.Deployment) (domain.Deployment, error) {
	if dep.TenantID == uuid.Nil || dep.Service == "" || dep.Environment == "" || dep.ImageTag == "" {
		return dep, domain.ErrInvalidArgument
	}
	if !domain.ValidStrategy(dep.Strategy) {
		dep.Strategy = "rolling"
	}
	if dep.ID == uuid.Nil {
		dep.ID = d.newID()
	}
	dep.Status = "started"
	dep.StartedAt = d.now()
	if d.GitOps != nil {
		if err := d.GitOps.Sync(ctx, dep.Environment, dep.Service, dep.ImageTag); err != nil {
			dep.Status = "failed"
			_ = d.Deployments.Save(ctx, dep)
			return dep, err
		}
	}
	if err := d.Deployments.Save(ctx, dep); err != nil {
		return dep, err
	}
	d.emit(ctx, dep.TenantID, dep.ID, domain.EventDeploymentStarted, map[string]any{
		"service": dep.Service, "environment": dep.Environment, "strategy": dep.Strategy, "imageTag": dep.ImageTag,
	})
	return dep, nil
}

func (d *Deps) CompleteDeployment(ctx context.Context, tenantID, id uuid.UUID, success bool) (domain.Deployment, error) {
	dep, err := d.Deployments.Get(ctx, tenantID, id)
	if err != nil {
		return dep, err
	}
	if dep.Status != "started" {
		return dep, domain.ErrIllegalTransition
	}
	now := d.now()
	dep.CompletedAt = &now
	if success {
		dep.Status = "succeeded"
		d.emit(ctx, tenantID, dep.ID, domain.EventDeploymentCompleted, map[string]any{
			"service": dep.Service, "environment": dep.Environment,
		})
	} else {
		dep.Status = "failed"
	}
	return dep, d.Deployments.Save(ctx, dep)
}

func (d *Deps) Rollback(ctx context.Context, tenantID, id uuid.UUID) (domain.Deployment, error) {
	dep, err := d.Deployments.Get(ctx, tenantID, id)
	if err != nil {
		return dep, err
	}
	if d.GitOps != nil {
		_ = d.GitOps.Rollback(ctx, dep.Environment, dep.Service)
	}
	now := d.now()
	dep.Status = "rolled_back"
	dep.CompletedAt = &now
	_ = d.Deployments.Save(ctx, dep)
	d.emit(ctx, tenantID, dep.ID, domain.EventRollbackTriggered, map[string]any{
		"service": dep.Service, "environment": dep.Environment,
	})
	return dep, nil
}

func (d *Deps) Scale(ctx context.Context, tenantID uuid.UUID, service, env string, from, to int, reason string) (domain.ScalingEvent, error) {
	if tenantID == uuid.Nil || service == "" || to < 0 {
		return domain.ScalingEvent{}, domain.ErrInvalidArgument
	}
	if d.Cluster != nil {
		if err := d.Cluster.Scale(ctx, env, service, to); err != nil {
			return domain.ScalingEvent{}, err
		}
	}
	ev := domain.ScalingEvent{
		ID: d.newID(), TenantID: tenantID, Service: service, Environment: env,
		FromReplicas: from, ToReplicas: to, Reason: reason, CreatedAt: d.now(),
	}
	_ = d.Scaling.Save(ctx, ev)
	d.emit(ctx, tenantID, ev.ID, domain.EventScalingTriggered, map[string]any{
		"service": service, "from": from, "to": to, "reason": reason,
	})
	return ev, nil
}

func (d *Deps) RunBackup(ctx context.Context, b domain.BackupJob) (domain.BackupJob, error) {
	if b.TenantID == uuid.Nil || b.Kind == "" || b.Target == "" {
		return b, domain.ErrInvalidArgument
	}
	if b.ID == uuid.Nil {
		b.ID = d.newID()
	}
	b.Status = "running"
	b.StartedAt = d.now()
	loc := "s3://nexora-backups/" + b.Kind + "/" + b.Target
	if d.BackupTool != nil {
		var err error
		loc, err = d.BackupTool.RunBackup(ctx, b.Kind, b.Target)
		if err != nil {
			b.Status = "failed"
			_ = d.Backups.Save(ctx, b)
			return b, err
		}
	}
	now := d.now()
	b.Location = loc
	b.Status = "completed"
	b.CompletedAt = &now
	_ = d.Backups.Save(ctx, b)
	d.emit(ctx, b.TenantID, b.ID, domain.EventBackupCompleted, map[string]any{
		"kind": b.Kind, "target": b.Target, "location": loc,
	})
	return b, nil
}

func (d *Deps) StartRecovery(ctx context.Context, r domain.RecoveryJob) (domain.RecoveryJob, error) {
	if r.TenantID == uuid.Nil || r.Kind == "" {
		return r, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		r.ID = d.newID()
	}
	r.Status = "started"
	r.StartedAt = d.now()
	_ = d.Recoveries.Save(ctx, r)
	d.emit(ctx, r.TenantID, r.ID, domain.EventRecoveryStarted, map[string]any{"kind": r.Kind})
	return r, nil
}

func (d *Deps) CompleteRecovery(ctx context.Context, tenantID, id uuid.UUID, notes string) (domain.RecoveryJob, error) {
	// list+match via save overwrite — get not on port; store by saving completed from StartRecovery path
	r := domain.RecoveryJob{ID: id, TenantID: tenantID, Status: "completed", Notes: notes}
	now := d.now()
	r.CompletedAt = &now
	r.StartedAt = now
	items, _ := d.Recoveries.List(ctx, tenantID)
	for _, it := range items {
		if it.ID == id {
			it.Status = "completed"
			it.Notes = notes
			it.CompletedAt = &now
			return it, d.Recoveries.Save(ctx, it)
		}
	}
	return r, domain.ErrNotFound
}

func (d *Deps) FireAlert(ctx context.Context, a domain.AlertEvent) (domain.AlertEvent, error) {
	if a.TenantID == uuid.Nil || a.Name == "" {
		return a, domain.ErrInvalidArgument
	}
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	if a.Severity == "" {
		a.Severity = "high"
	}
	a.Status = "firing"
	a.FiredAt = d.now()
	_ = d.Alerts.Save(ctx, a)
	d.emit(ctx, a.TenantID, a.ID, domain.EventAlertTriggered, map[string]any{
		"name": a.Name, "severity": a.Severity,
	})
	return a, nil
}

func (d *Deps) RecordCost(ctx context.Context, c domain.CostSnapshot) (domain.CostSnapshot, error) {
	if c.TenantID == uuid.Nil || c.Environment == "" || c.AmountMinor < 0 {
		return c, domain.ErrInvalidArgument
	}
	if c.ID == uuid.Nil {
		c.ID = d.newID()
	}
	if c.Currency == "" {
		c.Currency = "USD"
	}
	c.CreatedAt = d.now()
	return c, d.Costs.Save(ctx, c)
}

func (d *Deps) RecordSLO(ctx context.Context, s domain.SLOReport) (domain.SLOReport, error) {
	if s.TenantID == uuid.Nil || s.Service == "" {
		return s, domain.ErrInvalidArgument
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	if s.Window == "" {
		s.Window = "30d"
	}
	s.BudgetLeft = 1 - domain.BurnRate(s.Objective, s.Actual)
	if s.BudgetLeft < 0 {
		s.BudgetLeft = 0
	}
	s.CreatedAt = d.now()
	return s, d.SLOs.Save(ctx, s)
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	deps, _ := d.Deployments.List(ctx, tenantID, "")
	alerts, _ := d.Alerts.List(ctx, tenantID, "firing")
	backs, _ := d.Backups.List(ctx, tenantID)
	return map[string]any{
		"deployments": len(deps), "firingAlerts": len(alerts), "backups": len(backs),
		"tenantId": tenantID.String(),
	}, nil
}
