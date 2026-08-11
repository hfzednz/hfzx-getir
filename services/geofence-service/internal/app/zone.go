package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/domain"
)

// CreateZoneInput creates a new geofence zone.
type CreateZoneInput struct {
	TenantID  uuid.UUID
	Name      string
	City      string
	Kind      domain.ZoneKind
	Vertices  []domain.Point
	CenterLat *float64
	CenterLng *float64
	RadiusM   *float64
	Active    *bool
}

// CreateZone persists a new zone and emits ZoneChanged.
func (d *Deps) CreateZone(ctx context.Context, in CreateZoneInput) (domain.Zone, error) {
	if in.TenantID == uuid.Nil {
		return domain.Zone{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	z := domain.Zone{
		ID: d.newID(), TenantID: in.TenantID, Name: in.Name, City: in.City,
		Kind: in.Kind, Vertices: in.Vertices,
		CenterLat: in.CenterLat, CenterLng: in.CenterLng, RadiusM: in.RadiusM,
		Active: active, CreatedAt: now, UpdatedAt: now,
	}
	if err := z.Validate(); err != nil {
		return domain.Zone{}, err
	}
	if err := d.Zones.Create(ctx, z); err != nil {
		return domain.Zone{}, err
	}
	d.emit(ctx, z.TenantID, z.ID, domain.EventZoneChanged, map[string]any{
		"action": "created", "kind": string(z.Kind), "city": z.City,
	})
	return z, nil
}

// UpdateZoneInput updates an existing zone.
type UpdateZoneInput struct {
	TenantID  uuid.UUID
	ID        uuid.UUID
	Name      *string
	City      *string
	Kind      *domain.ZoneKind
	Vertices  []domain.Point
	CenterLat *float64
	CenterLng *float64
	RadiusM   *float64
	Active    *bool
	SetGeo    bool
}

// UpdateZone mutates a zone and emits ZoneChanged.
func (d *Deps) UpdateZone(ctx context.Context, in UpdateZoneInput) (domain.Zone, error) {
	z, err := d.Zones.Get(ctx, in.TenantID, in.ID)
	if err != nil {
		return domain.Zone{}, err
	}
	if in.Name != nil {
		z.Name = *in.Name
	}
	if in.City != nil {
		z.City = *in.City
	}
	if in.Kind != nil {
		z.Kind = *in.Kind
	}
	if in.Active != nil {
		z.Active = *in.Active
	}
	if in.SetGeo {
		z.Vertices = in.Vertices
		z.CenterLat = in.CenterLat
		z.CenterLng = in.CenterLng
		z.RadiusM = in.RadiusM
	}
	z.UpdatedAt = d.now()
	if err := z.Validate(); err != nil {
		return domain.Zone{}, err
	}
	if err := d.Zones.Update(ctx, z); err != nil {
		return domain.Zone{}, err
	}
	d.emit(ctx, z.TenantID, z.ID, domain.EventZoneChanged, map[string]any{
		"action": "updated", "kind": string(z.Kind), "city": z.City,
	})
	return z, nil
}

// DeleteZone soft-deletes by removing the zone row.
func (d *Deps) DeleteZone(ctx context.Context, tenantID, id uuid.UUID) error {
	z, err := d.Zones.Get(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if err := d.Zones.Delete(ctx, tenantID, id); err != nil {
		return err
	}
	d.emit(ctx, tenantID, id, domain.EventZoneChanged, map[string]any{
		"action": "deleted", "kind": string(z.Kind), "city": z.City,
	})
	return nil
}

// ContainsInput checks whether a point is inside a zone.
type ContainsInput struct {
	TenantID uuid.UUID
	ZoneID   uuid.UUID
	Point    domain.Point
}

// ContainsResult is the Contains outcome.
type ContainsResult struct {
	Inside bool
	ZoneID uuid.UUID
	Kind   domain.ZoneKind
}

// Contains checks a single zone.
func (d *Deps) Contains(ctx context.Context, in ContainsInput) (ContainsResult, error) {
	z, err := d.Zones.Get(ctx, in.TenantID, in.ZoneID)
	if err != nil {
		return ContainsResult{}, err
	}
	return ContainsResult{Inside: z.Contains(in.Point), ZoneID: z.ID, Kind: z.Kind}, nil
}

// ServiceabilityInput checks city+point coverage.
type ServiceabilityInput struct {
	TenantID uuid.UUID
	City     string
	Point    domain.Point
}

// CheckServiceability evaluates delivery coverage and restricted blocks.
func (d *Deps) CheckServiceability(ctx context.Context, in ServiceabilityInput) (domain.ServiceabilityResult, error) {
	if in.TenantID == uuid.Nil {
		return domain.ServiceabilityResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	zones, err := d.Zones.ListActive(ctx, in.TenantID, in.City)
	if err != nil {
		return domain.ServiceabilityResult{}, err
	}
	return domain.CheckServiceability(zones, in.City, in.Point), nil
}

// ListZonesInput filters zones.
type ListZonesInput struct {
	TenantID uuid.UUID
	City     string
	Kind     domain.ZoneKind
	Limit    int
	Offset   int
}

// ListZones returns paginated zones.
func (d *Deps) ListZones(ctx context.Context, in ListZonesInput) ([]domain.Zone, int, error) {
	if in.Limit <= 0 {
		in.Limit = 50
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	return d.Zones.List(ctx, in.TenantID, in.City, in.Kind, in.Limit, in.Offset)
}

// GetZone returns a zone by id.
func (d *Deps) GetZone(ctx context.Context, tenantID, id uuid.UUID) (domain.Zone, error) {
	return d.Zones.Get(ctx, tenantID, id)
}
