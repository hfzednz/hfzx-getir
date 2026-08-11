package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// UpsertForecastCmd upserts a forecast projection row.
type UpsertForecastCmd struct {
	TenantID        uuid.UUID
	WarehouseID     uuid.UUID
	VariantID       uuid.UUID
	SKUCode         string
	HorizonStart    time.Time
	HorizonEnd      time.Time
	PredictedDemand float64
	PredictedATP    *float64
	Confidence      *float64
	ModelID         string
	ModelVersion    string
}

// UpsertForecast stores an AI forecast projection.
func (d *Deps) UpsertForecast(ctx context.Context, in UpsertForecastCmd) (domain.StockForecast, error) {
	now := d.now()
	f := domain.StockForecast{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		VariantID: in.VariantID, SKUCode: in.SKUCode,
		HorizonStart: in.HorizonStart, HorizonEnd: in.HorizonEnd,
		PredictedDemand: in.PredictedDemand, PredictedATP: in.PredictedATP,
		Confidence: in.Confidence, ModelID: in.ModelID, ModelVersion: in.ModelVersion,
		Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	if err := f.Validate(); err != nil {
		return domain.StockForecast{}, err
	}
	if err := d.Forecasts.Upsert(ctx, f); err != nil {
		return domain.StockForecast{}, err
	}
	return f, nil
}

// GetForecast returns a forecast by id.
func (d *Deps) GetForecast(ctx context.Context, tenantID, id uuid.UUID) (domain.StockForecast, error) {
	return d.Forecasts.GetByID(ctx, tenantID, id)
}

// ListForecasts lists forecasts for warehouse/variant.
func (d *Deps) ListForecasts(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.StockForecast, error) {
	return d.Forecasts.List(ctx, tenantID, warehouseID, variantID)
}

// GenerateForecastCmd asks the stub AI port to produce a forecast and upserts it.
type GenerateForecastCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	HorizonDays int
}

// GenerateForecast uses the stub AI client then upserts.
func (d *Deps) GenerateForecast(ctx context.Context, in GenerateForecastCmd) (domain.StockForecast, error) {
	if d.AI == nil {
		return domain.StockForecast{}, fmt.Errorf("%w: AI forecast client not configured", domain.ErrInvalidArgument)
	}
	days := in.HorizonDays
	if days <= 0 {
		days = 7
	}
	pred, err := d.AI.Predict(ctx, in.TenantID, in.WarehouseID, in.VariantID, days)
	if err != nil {
		return domain.StockForecast{}, err
	}
	if pred.ID == uuid.Nil {
		pred.ID = d.newID()
	}
	pred.TenantID = in.TenantID
	pred.WarehouseID = in.WarehouseID
	pred.VariantID = in.VariantID
	now := d.now()
	if pred.HorizonStart.IsZero() {
		pred.HorizonStart = now
	}
	if pred.HorizonEnd.IsZero() {
		pred.HorizonEnd = now.AddDate(0, 0, days)
	}
	pred.UpdatedAt = now
	if pred.CreatedAt.IsZero() {
		pred.CreatedAt = now
	}
	if err := pred.Validate(); err != nil {
		return domain.StockForecast{}, err
	}
	if err := d.Forecasts.Upsert(ctx, pred); err != nil {
		return domain.StockForecast{}, err
	}
	return pred, nil
}
