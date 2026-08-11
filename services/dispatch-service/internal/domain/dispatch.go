package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// DispatchStatus is the delivery job lifecycle.
type DispatchStatus string

const (
	StatusQueued        DispatchStatus = "queued"
	StatusAssigned      DispatchStatus = "assigned"
	StatusPickupStarted DispatchStatus = "pickup_started"
	StatusPickedUp      DispatchStatus = "picked_up"
	StatusInTransit     DispatchStatus = "in_transit"
	StatusArrived       DispatchStatus = "arrived"
	StatusDelivered     DispatchStatus = "delivered"
	StatusFailed        DispatchStatus = "failed"
)

// Valid reports whether the status is recognized.
func (s DispatchStatus) Valid() bool {
	switch s {
	case StatusQueued, StatusAssigned, StatusPickupStarted, StatusPickedUp,
		StatusInTransit, StatusArrived, StatusDelivered, StatusFailed:
		return true
	default:
		return false
	}
}

// CanTransition reports whether from→to is allowed.
func CanTransition(from, to DispatchStatus) bool {
	switch from {
	case StatusQueued:
		return to == StatusAssigned
	case StatusAssigned:
		return to == StatusPickupStarted || to == StatusQueued || to == StatusFailed
	case StatusPickupStarted:
		return to == StatusPickedUp || to == StatusFailed
	case StatusPickedUp:
		return to == StatusInTransit || to == StatusFailed
	case StatusInTransit:
		return to == StatusArrived || to == StatusFailed
	case StatusArrived:
		return to == StatusDelivered || to == StatusFailed
	case StatusFailed:
		return to == StatusQueued || to == StatusAssigned
	default:
		return false
	}
}

// PODType is proof-of-delivery evidence kind.
type PODType string

const (
	PODOTP       PODType = "otp"
	PODQR        PODType = "qr"
	PODPhoto     PODType = "photo"
	PODSignature PODType = "signature"
	PODGPS       PODType = "gps"
)

// Valid reports whether the POD type is recognized.
func (t PODType) Valid() bool {
	switch t {
	case PODOTP, PODQR, PODPhoto, PODSignature, PODGPS:
		return true
	default:
		return false
	}
}

// FailReason classifies a failed delivery.
type FailReason string

const (
	FailCustomerUnavailable FailReason = "customer_unavailable"
	FailAddressIssue        FailReason = "address_issue"
	FailRefused             FailReason = "refused"
	FailDamaged             FailReason = "damaged"
	FailCourierIssue        FailReason = "courier_issue"
	FailOther               FailReason = "other"
)

// Valid reports whether the fail reason is recognized.
func (r FailReason) Valid() bool {
	switch r {
	case FailCustomerUnavailable, FailAddressIssue, FailRefused, FailDamaged, FailCourierIssue, FailOther:
		return true
	default:
		return false
	}
}

// VehicleType classifies fleet vehicles / courier capacity fit.
type VehicleType string

const (
	VehicleBike     VehicleType = "bike"
	VehicleScooter  VehicleType = "scooter"
	VehicleCar      VehicleType = "car"
	VehicleVan      VehicleType = "van"
)

// Valid reports whether the vehicle type is recognized.
func (v VehicleType) Valid() bool {
	switch v {
	case VehicleBike, VehicleScooter, VehicleCar, VehicleVan:
		return true
	default:
		return false
	}
}

// Point is a WGS84 coordinate.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// HaversineMeters returns great-circle distance in meters.
func HaversineMeters(a, b Point) float64 {
	const earth = 6371000.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earth * math.Asin(math.Min(1, math.Sqrt(h)))
}

// Dispatch is a delivery assignment job.
type Dispatch struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	OrderID             uuid.UUID // opaque
	FulfillmentID       uuid.UUID // opaque
	WarehouseID         uuid.UUID // opaque
	CourierPrincipalID  *uuid.UUID
	VehicleID           *uuid.UUID
	Status              DispatchStatus
	Pickup              Point
	Dropoff             Point
	RequiredVehicle     VehicleType
	BatchID             *uuid.UUID
	RouteID             *uuid.UUID
	ETASeconds          *int
	PODType             *PODType
	PODReference        string
	FailReason          *FailReason
	FailNote            string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Validate checks dispatch invariants.
func (d Dispatch) Validate() error {
	if d.ID == uuid.Nil {
		return fmt.Errorf("%w: dispatch id required", ErrInvalidArgument)
	}
	if d.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if d.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if !d.Status.Valid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidArgument, d.Status)
	}
	if d.RequiredVehicle != "" && !d.RequiredVehicle.Valid() {
		return fmt.Errorf("%w: invalid vehicle type %q", ErrInvalidArgument, d.RequiredVehicle)
	}
	return nil
}

// Transition applies a status change when legal.
func (d *Dispatch) Transition(to DispatchStatus) error {
	if !CanTransition(d.Status, to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, d.Status, to)
	}
	d.Status = to
	return nil
}

// DispatchEvent is an append-only status/event audit row.
type DispatchEvent struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	DispatchID uuid.UUID
	Type       string
	FromStatus DispatchStatus
	ToStatus   DispatchStatus
	Payload    map[string]any
	CreatedAt  time.Time
}

// CourierSnapshot is a courier availability snapshot used for assignment.
type CourierSnapshot struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	CourierPrincipalID uuid.UUID
	Available          bool
	Lat                float64
	Lng                float64
	CurrentLoad        int
	MaxCapacity        int
	Rating             float64
	VehicleType        VehicleType
	OnShift            bool
	UpdatedAt          time.Time
}

// HasCapacity reports whether the courier can take another job.
func (c CourierSnapshot) HasCapacity() bool {
	return c.Available && c.OnShift && c.CurrentLoad < c.MaxCapacity
}

// Vehicle is a fleet vehicle registry entry.
type Vehicle struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Plate     string
	Type      VehicleType
	Capacity  int
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks vehicle invariants.
func (v Vehicle) Validate() error {
	if v.ID == uuid.Nil {
		return fmt.Errorf("%w: vehicle id required", ErrInvalidArgument)
	}
	if v.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if v.Plate == "" {
		return fmt.Errorf("%w: plate required", ErrInvalidArgument)
	}
	if !v.Type.Valid() {
		return fmt.Errorf("%w: invalid vehicle type %q", ErrInvalidArgument, v.Type)
	}
	if v.Capacity < 0 {
		return fmt.Errorf("%w: capacity must be non-negative", ErrInvalidArgument)
	}
	return nil
}

// AssignmentAttempt records an auto/manual assign try.
type AssignmentAttempt struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	DispatchID         uuid.UUID
	CourierPrincipalID *uuid.UUID
	Strategy           string
	Success            bool
	DistanceM          *float64
	Reason             string
	CreatedAt          time.Time
}

// Batch groups multiple dispatches.
type Batch struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Label      string
	DispatchIDs []uuid.UUID
	CreatedAt  time.Time
}

// SelectNearestCourier picks the nearest available courier with capacity and vehicle fit.
func SelectNearestCourier(pickup Point, required VehicleType, pool []CourierSnapshot) (CourierSnapshot, float64, error) {
	var best CourierSnapshot
	bestDist := math.MaxFloat64
	found := false
	for _, c := range pool {
		if !c.HasCapacity() {
			continue
		}
		if required != "" && c.VehicleType != required {
			continue
		}
		d := HaversineMeters(pickup, Point{Lat: c.Lat, Lng: c.Lng})
		if d < bestDist {
			bestDist = d
			best = c
			found = true
		}
	}
	if !found {
		return CourierSnapshot{}, 0, ErrNoCourier
	}
	return best, bestDist, nil
}
