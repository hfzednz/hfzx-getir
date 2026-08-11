package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// POIKind classifies a spatial point of interest.
type POIKind string

const (
	POIKindWarehouse POIKind = "warehouse"
	POIKindPickup    POIKind = "pickup"
	POIKindPartner   POIKind = "partner"
	POIKindStore     POIKind = "store"
	POIKindDropoff   POIKind = "dropoff"
)

// Valid reports whether the kind is recognized.
func (k POIKind) Valid() bool {
	switch k {
	case POIKindWarehouse, POIKindPickup, POIKindPartner, POIKindStore, POIKindDropoff:
		return true
	default:
		return false
	}
}

// POI is a spatial index entry (opaque ref_id to other services).
type POI struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Kind      POIKind
	RefID     string
	Name      string
	Lat       float64
	Lng       float64
	Meta      map[string]any
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks POI invariants.
func (p POI) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: poi id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !p.Kind.Valid() {
		return fmt.Errorf("%w: invalid poi kind %q", ErrInvalidArgument, p.Kind)
	}
	if strings.TrimSpace(p.RefID) == "" {
		return fmt.Errorf("%w: ref_id required", ErrInvalidArgument)
	}
	if !ValidLatLng(p.Lat, p.Lng) {
		return fmt.Errorf("%w: lat/lng out of range", ErrInvalidArgument)
	}
	return nil
}

// NearbyQuery requests POIs near a center within a radius.
type NearbyQuery struct {
	TenantID uuid.UUID
	Center   LatLng
	RadiusM  float64
	Kind     *POIKind
	Limit    int
}

// Validate checks nearby query invariants.
func (q NearbyQuery) Validate() error {
	if q.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if err := q.Center.Validate(); err != nil {
		return err
	}
	if q.RadiusM <= 0 {
		return fmt.Errorf("%w: radius_m must be > 0", ErrInvalidArgument)
	}
	if q.Kind != nil && !q.Kind.Valid() {
		return fmt.Errorf("%w: invalid poi kind %q", ErrInvalidArgument, *q.Kind)
	}
	return nil
}
