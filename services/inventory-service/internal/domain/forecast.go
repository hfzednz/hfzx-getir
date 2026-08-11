package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StockForecast is an AI / demand-planning projection row (read model store).
// Not a write path for live stock mutations.
type StockForecast struct {
	ID              uuid.UUID
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
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks structural invariants.
func (f StockForecast) Validate() error {
	if f.ID == uuid.Nil {
		return fmt.Errorf("%w: forecast id required", ErrInvalidArgument)
	}
	if f.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if f.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if f.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if f.HorizonStart.IsZero() || f.HorizonEnd.IsZero() {
		return fmt.Errorf("%w: horizon_start/end required", ErrInvalidArgument)
	}
	if f.HorizonEnd.Before(f.HorizonStart) {
		return fmt.Errorf("%w: horizon_end before horizon_start", ErrInvariant)
	}
	if f.PredictedDemand < 0 {
		return fmt.Errorf("%w: predicted_demand cannot be negative", ErrInvalidArgument)
	}
	if f.Confidence != nil && (*f.Confidence < 0 || *f.Confidence > 1) {
		return fmt.Errorf("%w: confidence must be in [0,1]", ErrInvalidArgument)
	}
	return nil
}

// HorizonDays returns the inclusive calendar-day span of the horizon.
func (f StockForecast) HorizonDays() int {
	start := time.Date(f.HorizonStart.Year(), f.HorizonStart.Month(), f.HorizonStart.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(f.HorizonEnd.Year(), f.HorizonEnd.Month(), f.HorizonEnd.Day(), 0, 0, 0, 0, time.UTC)
	return int(end.Sub(start).Hours()/24) + 1
}

// Covers reports whether asOf falls within [horizon_start, horizon_end].
func (f StockForecast) Covers(asOf time.Time) bool {
	day := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	start := time.Date(f.HorizonStart.Year(), f.HorizonStart.Month(), f.HorizonStart.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(f.HorizonEnd.Year(), f.HorizonEnd.Month(), f.HorizonEnd.Day(), 0, 0, 0, 0, time.UTC)
	return !day.Before(start) && !day.After(end)
}
