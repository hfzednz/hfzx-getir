package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ZoneKind classifies a geofence zone.
type ZoneKind string

const (
	ZoneKindDelivery    ZoneKind = "delivery"
	ZoneKindRestricted  ZoneKind = "restricted"
	ZoneKindWarehouse   ZoneKind = "warehouse"
)

// Valid reports whether the zone kind is recognized.
func (k ZoneKind) Valid() bool {
	switch k {
	case ZoneKindDelivery, ZoneKindRestricted, ZoneKindWarehouse:
		return true
	default:
		return false
	}
}

// Zone is a delivery / restricted / warehouse geofence.
type Zone struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	City       string
	Kind       ZoneKind
	Vertices   []Point // polygon; empty when radius-only
	CenterLat  *float64
	CenterLng  *float64
	RadiusM    *float64
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks zone invariants.
func (z Zone) Validate() error {
	if z.ID == uuid.Nil {
		return fmt.Errorf("%w: zone id required", ErrInvalidArgument)
	}
	if z.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if z.Name == "" {
		return fmt.Errorf("%w: name required", ErrInvalidArgument)
	}
	if !z.Kind.Valid() {
		return fmt.Errorf("%w: invalid kind %q", ErrInvalidArgument, z.Kind)
	}
	hasPoly := len(z.Vertices) >= 3
	hasRadius := z.CenterLat != nil && z.CenterLng != nil && z.RadiusM != nil && *z.RadiusM > 0
	if !hasPoly && !hasRadius {
		return fmt.Errorf("%w: zone needs polygon (>=3 vertices) or radius geometry", ErrInvalidArgument)
	}
	if hasRadius && *z.RadiusM <= 0 {
		return fmt.Errorf("%w: radius must be positive", ErrInvalidArgument)
	}
	return nil
}

// Contains reports whether point is inside this zone (polygon and/or radius).
func (z Zone) Contains(p Point) bool {
	if !z.Active {
		return false
	}
	if len(z.Vertices) >= 3 {
		if PointInPolygon(p, z.Vertices) {
			return true
		}
	}
	if z.CenterLat != nil && z.CenterLng != nil && z.RadiusM != nil {
		return PointInRadius(p, Point{Lat: *z.CenterLat, Lng: *z.CenterLng}, *z.RadiusM)
	}
	return false
}

// ServiceabilityResult is the outcome of a city+point serviceability check.
type ServiceabilityResult struct {
	Serviceable      bool
	Blocked          bool
	Reason           string
	MatchingZones    []uuid.UUID
	RestrictedZones  []uuid.UUID
	DeliveryZones    []uuid.UUID
}

// CheckServiceability evaluates zones for a city and point.
// Restricted zones block; at least one active delivery zone must cover the point.
func CheckServiceability(zones []Zone, city string, p Point) ServiceabilityResult {
	res := ServiceabilityResult{
		MatchingZones:   make([]uuid.UUID, 0),
		RestrictedZones: make([]uuid.UUID, 0),
		DeliveryZones:   make([]uuid.UUID, 0),
	}
	for _, z := range zones {
		if !z.Active {
			continue
		}
		if city != "" && z.City != "" && z.City != city {
			continue
		}
		if !z.Contains(p) {
			continue
		}
		res.MatchingZones = append(res.MatchingZones, z.ID)
		switch z.Kind {
		case ZoneKindRestricted:
			res.RestrictedZones = append(res.RestrictedZones, z.ID)
		case ZoneKindDelivery:
			res.DeliveryZones = append(res.DeliveryZones, z.ID)
		}
	}
	if len(res.RestrictedZones) > 0 {
		res.Blocked = true
		res.Serviceable = false
		res.Reason = "restricted_zone"
		return res
	}
	if len(res.DeliveryZones) == 0 {
		res.Serviceable = false
		res.Reason = "outside_delivery_zone"
		return res
	}
	res.Serviceable = true
	res.Reason = "ok"
	return res
}
