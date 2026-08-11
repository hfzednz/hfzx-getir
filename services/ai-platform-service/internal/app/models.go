package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/domain"
)

// UpsertFeature writes online/offline feature values.
func (d *Deps) UpsertFeature(ctx context.Context, f domain.FeatureRecord) (domain.FeatureRecord, error) {
	if f.TenantID == uuid.Nil || f.EntityID == uuid.Nil || strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.EntityType) == "" {
		return f, domain.ErrInvalidArgument
	}
	if f.Version <= 0 {
		f.Version = 1
	}
	if f.Values == nil {
		f.Values = map[string]float64{}
	}
	f.UpdatedAt = d.now()
	if err := d.Features.Upsert(ctx, f); err != nil {
		return f, err
	}
	d.emit(ctx, f.TenantID, f.EntityID, domain.EventFeatureUpdated, map[string]any{
		"name": f.Name, "version": f.Version, "entityType": f.EntityType,
	})
	return f, nil
}

// RegisterModel creates/updates a model card.
func (d *Deps) RegisterModel(ctx context.Context, m domain.ModelCard) (domain.ModelCard, error) {
	if m.TenantID == uuid.Nil || m.Key == "" || m.Version == "" {
		return m, domain.ErrInvalidArgument
	}
	if m.Stage == "" {
		m.Stage = domain.StageDev
	}
	if !domain.ValidStage(m.Stage) {
		return m, domain.ErrInvalidArgument
	}
	if m.DeployStrat == "" {
		m.DeployStrat = domain.DeployStable
	}
	if m.ID == uuid.Nil {
		m.ID = d.newID()
	}
	now := d.now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	if err := d.Models.Save(ctx, m); err != nil {
		return m, err
	}
	d.emit(ctx, m.TenantID, m.ID, domain.EventModelTrained, map[string]any{
		"key": m.Key, "version": m.Version, "stage": m.Stage,
	})
	return m, nil
}

// PromoteModel moves a version across stages with optional approval.
func (d *Deps) PromoteModel(ctx context.Context, tenantID uuid.UUID, key, version, stage string, approver *uuid.UUID) (domain.ModelCard, error) {
	m, err := d.Models.Get(ctx, tenantID, key, version)
	if err != nil {
		return m, err
	}
	if !domain.ValidStage(stage) {
		return m, domain.ErrInvalidArgument
	}
	if stage == domain.StageProd && approver == nil {
		return m, domain.ErrForbidden
	}
	m.Stage = stage
	m.UpdatedAt = d.now()
	if stage == domain.StageProd {
		m.ApprovedBy = approver
		now := d.now()
		m.ApprovedAt = &now
	}
	if err := d.Models.Save(ctx, m); err != nil {
		return m, err
	}
	d.emit(ctx, tenantID, m.ID, domain.EventModelDeployed, map[string]any{
		"key": key, "version": version, "stage": stage, "strategy": m.DeployStrat,
	})
	return m, nil
}

// Infer runs real-time inference with canary/fallback routing.
func (d *Deps) Infer(ctx context.Context, req domain.InferenceRequest) (domain.InferenceResult, error) {
	var out domain.InferenceResult
	if req.TenantID == uuid.Nil || req.ModelKey == "" {
		return out, domain.ErrInvalidArgument
	}
	start := time.Now()

	var model domain.ModelCard
	var err error
	if req.Version != "" {
		model, err = d.Models.Get(ctx, req.TenantID, req.ModelKey, req.Version)
	} else {
		model, err = d.Models.GetProd(ctx, req.TenantID, req.ModelKey)
		if err != nil {
			// fallback to any registered version
			list, lerr := d.Models.List(ctx, req.TenantID, req.ModelKey)
			if lerr != nil || len(list) == 0 {
				return out, domain.ErrModelNotReady
			}
			model = list[0]
			err = nil
		}
	}
	if err != nil {
		return out, err
	}

	// Canary: route percentage of traffic to staging twin if configured
	useModel := model
	if model.DeployStrat == domain.DeployCanary && model.CanaryPct > 0 {
		if int(d.newID()[0])%100 < model.CanaryPct {
			if staging, err := findStage(ctx, d, req.TenantID, req.ModelKey, domain.StageStaging); err == nil {
				useModel = staging
			}
		}
	}

	feats := req.Features
	if feats == nil {
		feats = map[string]float64{}
	}
	// hydrate from feature store when entity provided
	if req.EntityID != nil && req.EntityType != "" {
		if rows, err := d.Features.ListByEntity(ctx, req.TenantID, req.EntityType, *req.EntityID); err == nil {
			for _, row := range rows {
				for k, v := range row.Values {
					if _, ok := feats[k]; !ok {
						feats[k] = v
					}
				}
			}
		}
	}

	preds, outputs, err := d.Runtime.Predict(ctx, useModel, feats, req.Inputs)
	if err != nil {
		if model.FallbackKey != "" {
			fb, ferr := d.Models.GetProd(ctx, req.TenantID, model.FallbackKey)
			if ferr == nil {
				preds, outputs, err = d.Runtime.Predict(ctx, fb, feats, req.Inputs)
				useModel = fb
			}
		}
		if err != nil {
			return out, err
		}
	}

	out = domain.InferenceResult{
		ID: d.newID(), ModelKey: useModel.Key, Version: useModel.Version, Stage: useModel.Stage,
		Predictions: preds, Outputs: outputs, LatencyMs: time.Since(start).Milliseconds(),
		Explain: domain.SimpleAttribution(feats), CreatedAt: d.now(),
	}

	// Shadow: also run prod twin without affecting response
	if model.Shadow || model.DeployStrat == domain.DeployShadow {
		if prod, err := d.Models.GetProd(ctx, req.TenantID, req.ModelKey); err == nil && prod.Version != useModel.Version {
			_, _, _ = d.Runtime.Predict(ctx, prod, feats, req.Inputs)
			out.Shadow = true
		}
	}

	d.emit(ctx, req.TenantID, out.ID, domain.EventInferenceCompleted, map[string]any{
		"modelKey": out.ModelKey, "version": out.Version, "latencyMs": out.LatencyMs,
	})
	d.emit(ctx, req.TenantID, out.ID, domain.EventPredictionGenerated, map[string]any{
		"predictions": out.Predictions,
	})
	return out, nil
}

func findStage(ctx context.Context, d *Deps, tenantID uuid.UUID, key, stage string) (domain.ModelCard, error) {
	list, err := d.Models.List(ctx, tenantID, key)
	if err != nil {
		return domain.ModelCard{}, err
	}
	for _, m := range list {
		if m.Stage == stage {
			return m, nil
		}
	}
	return domain.ModelCard{}, domain.ErrNotFound
}

// ForecastDemand convenience inference for demand forecasting.
func (d *Deps) ForecastDemand(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, horizonDays int) (domain.InferenceResult, error) {
	if horizonDays <= 0 {
		horizonDays = 7
	}
	return d.Infer(ctx, domain.InferenceRequest{
		TenantID: tenantID, ModelKey: "demand_forecast", EntityType: "product", EntityID: &productID,
		Features: map[string]float64{"horizon_days": float64(horizonDays)},
		Inputs:   map[string]any{"productId": productID.String()},
	})
}

// ScoreFraud convenience fraud inference.
func (d *Deps) ScoreFraud(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, features map[string]float64) (domain.InferenceResult, error) {
	return d.Infer(ctx, domain.InferenceRequest{
		TenantID: tenantID, ModelKey: "fraud_score", EntityType: entityType, EntityID: &entityID, Features: features,
	})
}

// SuggestPrice human-gated pricing suggestion.
func (d *Deps) SuggestPrice(ctx context.Context, tenantID, productID uuid.UUID, features map[string]float64) (domain.InferenceResult, error) {
	res, err := d.Infer(ctx, domain.InferenceRequest{
		TenantID: tenantID, ModelKey: "pricing_suggest", EntityType: "product", EntityID: &productID, Features: features,
	})
	if err != nil {
		return res, err
	}
	if res.Outputs == nil {
		res.Outputs = map[string]any{}
	}
	res.Outputs["humanGated"] = true
	return res, nil
}

// EmbedText exposes embedding for search/rec ports.
func (d *Deps) EmbedText(ctx context.Context, tenantID uuid.UUID, text string) ([]float64, error) {
	if d.Embed == nil {
		return nil, domain.ErrInvalidArgument
	}
	return d.Embed.Embed(ctx, tenantID, text)
}

// ReportDrift records monitoring drift.
func (d *Deps) ReportDrift(ctx context.Context, tenantID uuid.UUID, modelKey, metric string, value, threshold float64) (domain.DriftReport, error) {
	sev := "info"
	if value > threshold {
		sev = "warning"
	}
	if value > threshold*1.5 {
		sev = "critical"
	}
	rep := domain.DriftReport{
		ID: d.newID(), TenantID: tenantID, ModelKey: modelKey, Metric: metric,
		Value: value, Threshold: threshold, Severity: sev, CreatedAt: d.now(),
	}
	if err := d.Drift.Save(ctx, rep); err != nil {
		return rep, err
	}
	if sev != "info" {
		d.emit(ctx, tenantID, rep.ID, domain.EventDriftDetected, map[string]any{
			"modelKey": modelKey, "metric": metric, "value": value, "severity": sev,
		})
	}
	return rep, nil
}
