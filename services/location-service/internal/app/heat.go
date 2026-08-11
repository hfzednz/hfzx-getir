package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

// UpsertHeatCellInput upserts a demand heat cell.
type UpsertHeatCellInput struct {
	TenantID    uuid.UUID
	GridCell    string
	DemandScore float64
	Density     float64
}

// UpsertHeatCell stores a heat cell aggregate.
func (d *Deps) UpsertHeatCell(ctx context.Context, in UpsertHeatCellInput) (domain.HeatCell, error) {
	if in.TenantID == uuid.Nil {
		return domain.HeatCell{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	c := domain.HeatCell{
		ID: d.newID(), TenantID: in.TenantID,
		GridCell: strings.TrimSpace(in.GridCell),
		DemandScore: in.DemandScore, Density: in.Density, UpdatedAt: now,
	}
	// Stable id by reusing existing cell if repo returns by grid — memory upserts by grid key.
	if d.Heat == nil {
		return domain.HeatCell{}, fmt.Errorf("%w: heat repo not configured", domain.ErrInvariant)
	}
	if err := c.Validate(); err != nil {
		return domain.HeatCell{}, err
	}
	if err := d.Heat.Upsert(ctx, c); err != nil {
		return domain.HeatCell{}, err
	}
	return c, nil
}

// DemandHeatmapInput lists heat cells for a tenant.
type DemandHeatmapInput struct {
	TenantID uuid.UUID
	Limit    int
}

// DemandHeatmap returns demand heat cells.
func (d *Deps) DemandHeatmap(ctx context.Context, in DemandHeatmapInput) ([]domain.HeatCell, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if d.Heat == nil {
		return nil, fmt.Errorf("%w: heat repo not configured", domain.ErrInvariant)
	}
	return d.Heat.List(ctx, in.TenantID, in.Limit)
}

// GetOfflineManifestInput fetches an offline region manifest.
type GetOfflineManifestInput struct {
	TenantID uuid.UUID
	Region   string
}

// GetOfflineManifest returns offline package metadata (no tiles served).
func (d *Deps) GetOfflineManifest(ctx context.Context, in GetOfflineManifestInput) (domain.OfflineManifest, error) {
	if in.TenantID == uuid.Nil {
		return domain.OfflineManifest{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(in.Region) == "" {
		return domain.OfflineManifest{}, fmt.Errorf("%w: region required", domain.ErrInvalidArgument)
	}
	if d.Cache == nil {
		return domain.OfflineManifest{}, fmt.Errorf("%w: cache repo not configured", domain.ErrInvariant)
	}
	return d.Cache.GetOfflineManifest(ctx, in.TenantID, in.Region)
}

// UpsertOfflineManifestInput stores offline manifest metadata.
type UpsertOfflineManifestInput struct {
	TenantID  uuid.UUID
	Region    string
	Version   string
	URL       string
	SizeBytes int64
}

// UpsertOfflineManifest stores offline region metadata.
func (d *Deps) UpsertOfflineManifest(ctx context.Context, in UpsertOfflineManifestInput) (domain.OfflineManifest, error) {
	if in.TenantID == uuid.Nil {
		return domain.OfflineManifest{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	m := domain.OfflineManifest{
		ID: d.newID(), TenantID: in.TenantID,
		Region: strings.TrimSpace(in.Region), Version: strings.TrimSpace(in.Version),
		URL: strings.TrimSpace(in.URL), SizeBytes: in.SizeBytes, UpdatedAt: d.now(),
	}
	if err := m.Validate(); err != nil {
		return domain.OfflineManifest{}, err
	}
	if d.Cache == nil {
		return domain.OfflineManifest{}, fmt.Errorf("%w: cache repo not configured", domain.ErrInvariant)
	}
	if err := d.Cache.UpsertOfflineManifest(ctx, m); err != nil {
		return domain.OfflineManifest{}, err
	}
	return m, nil
}
