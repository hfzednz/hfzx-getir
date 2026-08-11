package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

// CoverageStats is an admin coverage dashboard aggregate.
type CoverageStats struct {
	TenantID      uuid.UUID `json:"tenantId"`
	POICount      int       `json:"poiCount"`
	HeatCellCount int       `json:"heatCellCount"`
	AvgDemand     float64   `json:"avgDemand"`
	AvgDensity    float64   `json:"avgDensity"`
}

// AdminCoverageStatsInput requests coverage aggregates.
type AdminCoverageStatsInput struct {
	TenantID uuid.UUID
}

// AdminCoverageStats returns POI + heatmap coverage metrics.
func (d *Deps) AdminCoverageStats(ctx context.Context, in AdminCoverageStatsInput) (CoverageStats, error) {
	if in.TenantID == uuid.Nil {
		return CoverageStats{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	stats := CoverageStats{TenantID: in.TenantID}
	if d.POIs != nil {
		n, err := d.POIs.Count(ctx, in.TenantID)
		if err != nil {
			return CoverageStats{}, err
		}
		stats.POICount = n
	}
	if d.Heat != nil {
		cells, err := d.Heat.List(ctx, in.TenantID, 1000)
		if err != nil {
			return CoverageStats{}, err
		}
		stats.HeatCellCount = len(cells)
		if len(cells) > 0 {
			var sumD, sumN float64
			for _, c := range cells {
				sumD += c.DemandScore
				sumN += c.Density
			}
			stats.AvgDemand = sumD / float64(len(cells))
			stats.AvgDensity = sumN / float64(len(cells))
		}
	}
	return stats, nil
}
