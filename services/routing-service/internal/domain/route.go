package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// WaypointKind classifies a stop on a route.
type WaypointKind string

const (
	WaypointWarehouse WaypointKind = "warehouse"
	WaypointStop      WaypointKind = "stop"
	WaypointCourier   WaypointKind = "courier"
)

// Valid reports whether the kind is recognized.
func (k WaypointKind) Valid() bool {
	switch k {
	case WaypointWarehouse, WaypointStop, WaypointCourier:
		return true
	default:
		return false
	}
}

// RouteStatus is the lifecycle of a route plan.
type RouteStatus string

const (
	RouteStatusDraft     RouteStatus = "draft"
	RouteStatusOptimized RouteStatus = "optimized"
	RouteStatusActive    RouteStatus = "active"
	RouteStatusCompleted RouteStatus = "completed"
	RouteStatusCancelled RouteStatus = "cancelled"
)

// Valid reports whether the status is recognized.
func (s RouteStatus) Valid() bool {
	switch s {
	case RouteStatusDraft, RouteStatusOptimized, RouteStatusActive, RouteStatusCompleted, RouteStatusCancelled:
		return true
	default:
		return false
	}
}

// Waypoint is an ordered point on a route (warehouse → stops).
type Waypoint struct {
	ID       uuid.UUID
	Sequence int
	Kind     WaypointKind
	Lat      float64
	Lon      float64
	OrderID  *uuid.UUID // opaque OMS ref
	Label    string
	ETAAt    *time.Time
}

// Validate checks waypoint invariants.
func (w Waypoint) Validate() error {
	if !w.Kind.Valid() {
		return fmt.Errorf("%w: invalid waypoint kind %q", ErrInvalidArgument, w.Kind)
	}
	if !ValidLatLon(w.Lat, w.Lon) {
		return fmt.Errorf("%w: invalid lat/lon", ErrInvalidArgument)
	}
	if w.Sequence < 0 {
		return fmt.Errorf("%w: sequence must be >= 0", ErrInvalidArgument)
	}
	return nil
}

// RouteLeg is a segment between consecutive waypoints.
type RouteLeg struct {
	ID              uuid.UUID
	RouteID         uuid.UUID
	FromSequence    int
	ToSequence      int
	DistanceMeters  float64
	DurationSeconds float64
}

// Route is an ordered multi-stop plan with ETA factors.
type Route struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	DispatchID      *uuid.UUID // opaque dispatch ref
	CourierID       *uuid.UUID // opaque courier principal
	WarehouseID     *uuid.UUID // opaque warehouse ref
	Status          RouteStatus
	Waypoints       []Waypoint
	Legs            []RouteLeg
	DistanceMeters  float64
	DurationSeconds float64
	ETAAt           *time.Time
	TrafficFactor   float64
	WeatherFactor   float64
	SpeedMPS        float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks route invariants.
func (r Route) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: route id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !r.Status.Valid() {
		return fmt.Errorf("%w: invalid route status %q", ErrInvalidArgument, r.Status)
	}
	if len(r.Waypoints) == 0 {
		return fmt.Errorf("%w: at least one waypoint required", ErrInvalidArgument)
	}
	for _, w := range r.Waypoints {
		if err := w.Validate(); err != nil {
			return err
		}
	}
	if r.TrafficFactor < 0 {
		return fmt.Errorf("%w: traffic_factor must be >= 0", ErrInvalidArgument)
	}
	if r.WeatherFactor < 0 {
		return fmt.Errorf("%w: weather_factor must be >= 0", ErrInvalidArgument)
	}
	return nil
}

// ETASnapshot captures a point-in-time ETA estimate.
type ETASnapshot struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	RouteID         uuid.UUID
	DistanceMeters  float64
	DurationSeconds float64
	ETAAt           time.Time
	TrafficFactor   float64
	WeatherFactor   float64
	SpeedMPS        float64
	Reason          string
	CapturedAt      time.Time
}

// TrafficHint is a region traffic multiplier.
type TrafficHint struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	RegionKey  string
	Lat        float64
	Lon        float64
	RadiusM    float64
	Factor     float64
	ValidFrom  time.Time
	ValidUntil time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks traffic hint invariants.
func (h TrafficHint) Validate() error {
	if h.ID == uuid.Nil {
		return fmt.Errorf("%w: traffic hint id required", ErrInvalidArgument)
	}
	if h.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if h.Factor <= 0 {
		return fmt.Errorf("%w: factor must be > 0", ErrInvalidArgument)
	}
	if !ValidLatLon(h.Lat, h.Lon) {
		return fmt.Errorf("%w: invalid lat/lon", ErrInvalidArgument)
	}
	return nil
}

// OptimizeNearestNeighbor reorders stops after the origin using nearest-neighbor.
// The first waypoint (index 0) is treated as the fixed origin (warehouse/courier).
func OptimizeNearestNeighbor(waypoints []Waypoint) []Waypoint {
	if len(waypoints) <= 2 {
		out := make([]Waypoint, len(waypoints))
		copy(out, waypoints)
		for i := range out {
			out[i].Sequence = i
		}
		return out
	}

	origin := waypoints[0]
	remaining := make([]Waypoint, len(waypoints)-1)
	copy(remaining, waypoints[1:])

	ordered := make([]Waypoint, 0, len(waypoints))
	ordered = append(ordered, origin)
	curLat, curLon := origin.Lat, origin.Lon

	for len(remaining) > 0 {
		bestIdx := 0
		bestDist := math.MaxFloat64
		for i, w := range remaining {
			d := HaversineMeters(curLat, curLon, w.Lat, w.Lon)
			if d < bestDist {
				bestDist = d
				bestIdx = i
			}
		}
		next := remaining[bestIdx]
		ordered = append(ordered, next)
		curLat, curLon = next.Lat, next.Lon
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	for i := range ordered {
		ordered[i].Sequence = i
	}
	return ordered
}
