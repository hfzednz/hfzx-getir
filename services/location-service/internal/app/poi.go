package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

// UpsertPOIInput creates or updates a POI.
type UpsertPOIInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	Kind     domain.POIKind
	RefID    string
	Name     string
	Lat      float64
	Lng      float64
	Meta     map[string]any
	Active   *bool
}

// UpsertPOI persists a spatial POI.
func (d *Deps) UpsertPOI(ctx context.Context, in UpsertPOIInput) (domain.POI, error) {
	if in.TenantID == uuid.Nil {
		return domain.POI{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	id := in.ID
	if id == uuid.Nil {
		id = d.newID()
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	p := domain.POI{
		ID: id, TenantID: in.TenantID, Kind: in.Kind,
		RefID: strings.TrimSpace(in.RefID), Name: strings.TrimSpace(in.Name),
		Lat: in.Lat, Lng: in.Lng, Meta: in.Meta, Active: active,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := p.Validate(); err != nil {
		return domain.POI{}, err
	}
	if d.POIs == nil {
		return domain.POI{}, fmt.Errorf("%w: poi repo not configured", domain.ErrInvariant)
	}
	if err := d.POIs.Upsert(ctx, p); err != nil {
		return domain.POI{}, err
	}
	if d.Search != nil {
		if err := d.Search.IndexPOI(ctx, p); err != nil {
			// Non-fatal: Redis/Postgres remain source of truth for nearby.
			_ = err
		}
	}
	return p, nil
}

// NearbyInput finds POIs near a point.
type NearbyInput struct {
	TenantID uuid.UUID
	Lat      float64
	Lng      float64
	RadiusM  float64
	Kind     *domain.POIKind
	Limit    int
}

// Nearby returns POIs within radius (Haversine).
func (d *Deps) Nearby(ctx context.Context, in NearbyInput) ([]domain.POI, error) {
	q := domain.NearbyQuery{
		TenantID: in.TenantID,
		Center:   domain.LatLng{Lat: in.Lat, Lng: in.Lng},
		RadiusM:  in.RadiusM,
		Kind:     in.Kind,
		Limit:    in.Limit,
	}
	if err := q.Validate(); err != nil {
		return nil, err
	}
	if d.POIs == nil {
		return nil, fmt.Errorf("%w: poi repo not configured", domain.ErrInvariant)
	}
	return d.POIs.Nearby(ctx, q)
}

// NearestOfKindInput finds nearest POIs of a kind.
type NearestOfKindInput struct {
	TenantID uuid.UUID
	Kind     domain.POIKind
	Lat      float64
	Lng      float64
	Limit    int
}

// NearestOfKind returns closest POIs of the given kind.
func (d *Deps) NearestOfKind(ctx context.Context, in NearestOfKindInput) ([]domain.POI, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if !in.Kind.Valid() {
		return nil, fmt.Errorf("%w: invalid poi kind %q", domain.ErrInvalidArgument, in.Kind)
	}
	if !domain.ValidLatLng(in.Lat, in.Lng) {
		return nil, fmt.Errorf("%w: lat/lng out of range", domain.ErrInvalidArgument)
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 1
	}
	if d.POIs == nil {
		return nil, fmt.Errorf("%w: poi repo not configured", domain.ErrInvariant)
	}
	return d.POIs.NearestOfKind(ctx, in.TenantID, in.Kind, in.Lat, in.Lng, limit)
}

// RadiusSearchInput is an alias-style radius search.
type RadiusSearchInput struct {
	TenantID uuid.UUID
	Lat      float64
	Lng      float64
	RadiusM  float64
	Kind     *domain.POIKind
	Limit    int
}

// RadiusSearch finds POIs within a radius.
func (d *Deps) RadiusSearch(ctx context.Context, in RadiusSearchInput) ([]domain.POI, error) {
	return d.Nearby(ctx, NearbyInput(in))
}

// BBoxSearchInput finds POIs inside a bounding box.
type BBoxSearchInput struct {
	TenantID uuid.UUID
	BBox     domain.BBox
	Kind     *domain.POIKind
	Limit    int
}

// BBoxSearch returns POIs inside the bbox.
func (d *Deps) BBoxSearch(ctx context.Context, in BBoxSearchInput) ([]domain.POI, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if err := in.BBox.Validate(); err != nil {
		return nil, err
	}
	if d.POIs == nil {
		return nil, fmt.Errorf("%w: poi repo not configured", domain.ErrInvariant)
	}
	return d.POIs.InBBox(ctx, in.TenantID, in.BBox, in.Kind, in.Limit)
}
