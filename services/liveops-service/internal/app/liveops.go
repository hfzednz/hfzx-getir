package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/domain"
)

func (d *Deps) UpsertFlag(ctx context.Context, f domain.FeatureFlag) (domain.FeatureFlag, error) {
	if f.TenantID == uuid.Nil || f.Key == "" {
		return f, domain.ErrInvalidArgument
	}
	f.Key = domain.NormalizeKey(f.Key)
	now := d.now()
	existing, err := d.Flags.GetByKey(ctx, f.TenantID, f.Key)
	created := false
	if err != nil {
		if f.ID == uuid.Nil {
			f.ID = d.newID()
		}
		f.Version = 1
		f.CreatedAt = now
		created = true
	} else {
		f.ID = existing.ID
		f.Version = existing.Version + 1
		f.CreatedAt = existing.CreatedAt
	}
	if f.Percentage < 0 || f.Percentage > 100 {
		return f, domain.ErrInvalidArgument
	}
	f.UpdatedAt = now
	if err := d.Flags.Save(ctx, f); err != nil {
		return f, err
	}
	if created {
		d.emit(ctx, f.TenantID, f.ID, domain.EventFeatureFlagCreated, map[string]any{"key": f.Key})
	} else {
		d.emit(ctx, f.TenantID, f.ID, domain.EventFeatureFlagUpdated, map[string]any{"key": f.Key, "version": f.Version})
	}
	if f.Enabled && !f.EmergencyOff {
		d.emit(ctx, f.TenantID, f.ID, domain.EventFeatureEnabled, map[string]any{"key": f.Key})
	} else {
		d.emit(ctx, f.TenantID, f.ID, domain.EventFeatureDisabled, map[string]any{"key": f.Key})
	}
	return f, nil
}

func (d *Deps) EvaluateFlags(ctx context.Context, tenantID uuid.UUID, keys []string, eval domain.EvalContext) ([]domain.FlagEvaluation, error) {
	if tenantID == uuid.Nil || eval.SubjectID == "" {
		return nil, domain.ErrInvalidArgument
	}
	all, err := d.Flags.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	byKey := map[string]domain.FeatureFlag{}
	for _, f := range all {
		byKey[f.Key] = f
	}
	// resolve dependencies iteratively
	depOn := map[string]bool{}
	for _, f := range all {
		ev := domain.EvaluateFlag(f, eval, depOn)
		depOn[f.Key] = ev.Enabled
	}
	// second pass with deps known
	for _, f := range all {
		ev := domain.EvaluateFlag(f, eval, depOn)
		depOn[f.Key] = ev.Enabled
	}
	out := []domain.FlagEvaluation{}
	want := map[string]bool{}
	for _, k := range keys {
		want[domain.NormalizeKey(k)] = true
	}
	for _, f := range all {
		if len(want) > 0 && !want[f.Key] {
			continue
		}
		ev := domain.EvaluateFlag(f, eval, depOn)
		out = append(out, ev)
		if d.Cache != nil {
			b, _ := json.Marshal(ev)
			_ = d.Cache.Set(ctx, cacheKey(tenantID, f.Key, eval.SubjectID), string(b), 30*time.Second)
		}
	}
	return out, nil
}

func cacheKey(tenantID uuid.UUID, flagKey, subject string) string {
	return tenantID.String() + "|" + flagKey + "|" + subject
}

func (d *Deps) PublishConfig(ctx context.Context, c domain.ConfigDocument) (domain.ConfigDocument, error) {
	if c.TenantID == uuid.Nil || c.Key == "" || c.Namespace == "" {
		return c, domain.ErrInvalidArgument
	}
	c.Key = domain.NormalizeKey(c.Key)
	now := d.now()
	existing, err := d.Configs.GetByKey(ctx, c.TenantID, c.Key)
	if err != nil {
		if c.ID == uuid.Nil {
			c.ID = d.newID()
		}
		c.Version = 1
		c.CreatedAt = now
	} else {
		c.ID = existing.ID
		c.Version = existing.Version + 1
		c.CreatedAt = existing.CreatedAt
	}
	c.Status = "published"
	c.UpdatedAt = now
	if c.Payload == nil {
		c.Payload = map[string]any{}
	}
	if err := d.Configs.Save(ctx, c); err != nil {
		return c, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventConfigurationUpdated, map[string]any{
		"key": c.Key, "namespace": c.Namespace, "version": c.Version,
	})
	return c, nil
}

func (d *Deps) ResolveConfig(ctx context.Context, tenantID uuid.UUID, namespace string, eval domain.EvalContext) (map[string]any, error) {
	items, err := d.Configs.List(ctx, tenantID, namespace)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for _, c := range items {
		if c.Status != "published" {
			continue
		}
		out[c.Key] = c.Payload
	}
	// overlay active liveops events
	events, _ := d.Events.List(ctx, tenantID, "")
	now := d.now()
	for _, e := range events {
		if domain.EventActive(e, now, eval) && e.ConfigPatch != nil {
			for k, v := range e.ConfigPatch {
				out[k] = v
			}
		}
	}
	return out, nil
}

func (d *Deps) UpsertExperiment(ctx context.Context, e domain.Experiment) (domain.Experiment, error) {
	if e.TenantID == uuid.Nil || e.Key == "" || !domain.ValidExperimentKind(e.Kind) || len(e.Variants) < 2 {
		return e, domain.ErrInvalidArgument
	}
	e.Key = domain.NormalizeKey(e.Key)
	now := d.now()
	existing, err := d.Experiments.GetByKey(ctx, e.TenantID, e.Key)
	if err != nil {
		if e.ID == uuid.Nil {
			e.ID = d.newID()
		}
		e.CreatedAt = now
		if e.Status == "" {
			e.Status = "draft"
		}
	} else {
		e.ID = existing.ID
		e.CreatedAt = existing.CreatedAt
		if e.Status == "" {
			e.Status = existing.Status
		}
	}
	return e, d.Experiments.Save(ctx, e)
}

func (d *Deps) StartExperiment(ctx context.Context, tenantID uuid.UUID, key string) (domain.Experiment, error) {
	e, err := d.Experiments.GetByKey(ctx, tenantID, domain.NormalizeKey(key))
	if err != nil {
		return e, err
	}
	if e.Status != "draft" && e.Status != "paused" {
		return e, domain.ErrIllegalTransition
	}
	now := d.now()
	e.Status = "running"
	e.StartedAt = &now
	if err := d.Experiments.Save(ctx, e); err != nil {
		return e, err
	}
	d.emit(ctx, tenantID, e.ID, domain.EventExperimentStarted, map[string]any{"key": e.Key, "kind": e.Kind})
	return e, nil
}

func (d *Deps) AssignExperiment(ctx context.Context, tenantID uuid.UUID, key, subjectID string) (domain.Assignment, error) {
	if subjectID == "" {
		return domain.Assignment{}, domain.ErrInvalidArgument
	}
	e, err := d.Experiments.GetByKey(ctx, tenantID, domain.NormalizeKey(key))
	if err != nil {
		return domain.Assignment{}, err
	}
	if existing, ok, err := d.Experiments.GetAssignment(ctx, tenantID, e.ID, subjectID); err != nil {
		return domain.Assignment{}, err
	} else if ok {
		return existing, nil
	}
	variant, err := domain.AssignVariant(e, subjectID)
	if err != nil {
		return domain.Assignment{}, err
	}
	a := domain.Assignment{
		ID: d.newID(), TenantID: tenantID, ExperimentID: e.ID,
		SubjectID: subjectID, VariantKey: variant, AssignedAt: d.now(),
	}
	return a, d.Experiments.SaveAssignment(ctx, a)
}

func (d *Deps) CompleteExperiment(ctx context.Context, tenantID uuid.UUID, key string, rates map[string]float64, autoRollout bool) (domain.Experiment, error) {
	e, err := d.Experiments.GetByKey(ctx, tenantID, domain.NormalizeKey(key))
	if err != nil {
		return e, err
	}
	if e.Status != "running" && e.Status != "paused" {
		return e, domain.ErrIllegalTransition
	}
	keys := make([]string, 0, len(e.Variants))
	for _, v := range e.Variants {
		keys = append(keys, v.Key)
	}
	winner := domain.PickWinner(keys, rates)
	if d.AI != nil {
		if w, err := d.AI.SuggestWinner(ctx, tenantID, e.Key, rates); err == nil && w != "" {
			winner = w
		}
	}
	control := rates["control"]
	if control == 0 && len(keys) > 0 {
		control = rates[keys[0]]
	}
	if !domain.Significant(control, rates[winner], 0.02) && e.Kind != "aa" {
		// still complete but mark no auto
		autoRollout = false
	}
	now := d.now()
	e.Status = "completed"
	e.EndedAt = &now
	e.Winner = winner
	if err := d.Experiments.Save(ctx, e); err != nil {
		return e, err
	}
	d.emit(ctx, tenantID, e.ID, domain.EventExperimentCompleted, map[string]any{
		"key": e.Key, "winner": winner, "autoRollout": autoRollout,
	})
	if d.Metrics != nil {
		_ = d.Metrics.Ingest(ctx, tenantID, "experiment_completed", 1, map[string]string{"key": e.Key, "winner": winner})
	}
	if autoRollout && winner != "" {
		_, _ = d.UpsertFlag(ctx, domain.FeatureFlag{
			TenantID: tenantID, Key: "exp_" + e.Key + "_winner", Enabled: true, Percentage: 100,
			Description: "auto rollout from experiment " + e.Key,
		})
	}
	return e, nil
}

func (d *Deps) UpsertLiveEvent(ctx context.Context, e domain.LiveOpsEvent) (domain.LiveOpsEvent, error) {
	if e.TenantID == uuid.Nil || e.Key == "" || e.Kind == "" {
		return e, domain.ErrInvalidArgument
	}
	e.Key = domain.NormalizeKey(e.Key)
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.Status == "" {
		e.Status = "scheduled"
	}
	e.CreatedAt = d.now()
	return e, d.Events.Save(ctx, e)
}

func (d *Deps) ActivateEvent(ctx context.Context, tenantID uuid.UUID, key string) (domain.LiveOpsEvent, error) {
	e, err := d.Events.GetByKey(ctx, tenantID, domain.NormalizeKey(key))
	if err != nil {
		return e, err
	}
	e.Status = "active"
	return e, d.Events.Save(ctx, e)
}

func (d *Deps) RequestChange(ctx context.Context, c domain.ChangeRequest) (domain.ChangeRequest, error) {
	if c.TenantID == uuid.Nil || c.Kind == "" || c.SubjectKey == "" {
		return c, domain.ErrInvalidArgument
	}
	if c.ID == uuid.Nil {
		c.ID = d.newID()
	}
	c.Status = "pending"
	c.CreatedAt = d.now()
	return c, d.Changes.Save(ctx, c)
}

func (d *Deps) DecideChange(ctx context.Context, tenantID, id uuid.UUID, approve bool) (domain.ChangeRequest, error) {
	c, err := d.Changes.Get(ctx, tenantID, id)
	if err != nil {
		return c, err
	}
	if c.Status != "pending" {
		return c, domain.ErrIllegalTransition
	}
	now := d.now()
	c.DecidedAt = &now
	if approve {
		c.Status = "approved"
	} else {
		c.Status = "rejected"
	}
	return c, d.Changes.Save(ctx, c)
}

func (d *Deps) Rollback(ctx context.Context, tenantID uuid.UUID, kind, subjectKey, reason string) (domain.RollbackRecord, error) {
	if tenantID == uuid.Nil || kind == "" || subjectKey == "" {
		return domain.RollbackRecord{}, domain.ErrInvalidArgument
	}
	subjectKey = domain.NormalizeKey(subjectKey)
	rec := domain.RollbackRecord{
		ID: d.newID(), TenantID: tenantID, Kind: kind, SubjectKey: subjectKey,
		Reason: reason, CreatedAt: d.now(),
	}
	switch kind {
	case "flag":
		f, err := d.Flags.GetByKey(ctx, tenantID, subjectKey)
		if err != nil {
			return rec, err
		}
		rec.FromVersion = f.Version
		f.EmergencyOff = true
		f.Enabled = false
		f.Version++
		rec.ToVersion = f.Version
		_ = d.Flags.Save(ctx, f)
		d.emit(ctx, tenantID, f.ID, domain.EventFeatureDisabled, map[string]any{"key": f.Key, "emergency": true})
	case "config":
		c, err := d.Configs.GetByKey(ctx, tenantID, subjectKey)
		if err != nil {
			return rec, err
		}
		rec.FromVersion = c.Version
		c.Status = "archived"
		c.Version++
		rec.ToVersion = c.Version
		_ = d.Configs.Save(ctx, c)
	case "experiment":
		e, err := d.Experiments.GetByKey(ctx, tenantID, subjectKey)
		if err != nil {
			return rec, err
		}
		e.Status = "rolled_back"
		now := d.now()
		e.EndedAt = &now
		_ = d.Experiments.Save(ctx, e)
	case "emergency":
		flags, _ := d.Flags.List(ctx, tenantID)
		for _, f := range flags {
			f.EmergencyOff = true
			f.Enabled = false
			f.Version++
			_ = d.Flags.Save(ctx, f)
		}
	default:
		return rec, domain.ErrInvalidArgument
	}
	if err := d.Rollbacks.Save(ctx, rec); err != nil {
		return rec, err
	}
	d.emit(ctx, tenantID, rec.ID, domain.EventRollbackExecuted, map[string]any{
		"kind": kind, "subjectKey": subjectKey, "reason": reason,
	})
	return rec, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	flags, _ := d.Flags.List(ctx, tenantID)
	exps, _ := d.Experiments.List(ctx, tenantID)
	pending, _ := d.Changes.ListPending(ctx, tenantID)
	running := 0
	for _, e := range exps {
		if e.Status == "running" {
			running++
		}
	}
	enabled := 0
	for _, f := range flags {
		if f.Enabled && !f.EmergencyOff {
			enabled++
		}
	}
	return map[string]any{
		"flags": len(flags), "flagsEnabled": enabled, "experimentsRunning": running,
		"pendingApprovals": len(pending), "tenantId": tenantID.String(),
	}, nil
}
