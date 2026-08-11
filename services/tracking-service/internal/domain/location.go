package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LocationUpdate is an inbound GPS ping from a courier app.
type LocationUpdate struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CourierID  uuid.UUID // opaque courier principal
	Lat        float64
	Lon        float64
	AccuracyM  float64
	HeadingDeg *float64
	SpeedMPS   *float64
	RecordedAt time.Time
	ReceivedAt time.Time
}

// Validate checks location update invariants.
func (u LocationUpdate) Validate() error {
	if u.ID == uuid.Nil {
		return fmt.Errorf("%w: location id required", ErrInvalidArgument)
	}
	if u.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if u.CourierID == uuid.Nil {
		return fmt.Errorf("%w: courier_id required", ErrInvalidArgument)
	}
	if !ValidLatLon(u.Lat, u.Lon) {
		return fmt.Errorf("%w: invalid lat/lon", ErrInvalidArgument)
	}
	return nil
}

// CourierLocation is the latest known position for a courier.
type CourierLocation struct {
	TenantID   uuid.UUID
	CourierID  uuid.UUID
	Lat        float64
	Lon        float64
	AccuracyM  float64
	HeadingDeg *float64
	SpeedMPS   *float64
	RecordedAt time.Time
	UpdatedAt  time.Time
}

// LocationHistoryEntry is a capped historical ping.
type LocationHistoryEntry struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CourierID  uuid.UUID
	Lat        float64
	Lon        float64
	AccuracyM  float64
	HeadingDeg *float64
	SpeedMPS   *float64
	RecordedAt time.Time
	ReceivedAt time.Time
}

// TimelineEventType classifies delivery timeline projections.
type TimelineEventType string

const (
	TimelineLocationUpdated TimelineEventType = "LocationUpdated"
	TimelineArrived         TimelineEventType = "Arrived"
	TimelineGeofenceEnter   TimelineEventType = "GeofenceEnter"
	TimelineGeofenceExit    TimelineEventType = "GeofenceExit"
	TimelineCustom          TimelineEventType = "Custom"
)

// Valid reports whether the timeline event type is recognized.
func (t TimelineEventType) Valid() bool {
	switch t {
	case TimelineLocationUpdated, TimelineArrived, TimelineGeofenceEnter, TimelineGeofenceExit, TimelineCustom:
		return true
	default:
		return false
	}
}

// TimelineEvent is a delivery timeline projection row.
type TimelineEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	OrderID   uuid.UUID // opaque OMS ref
	CourierID *uuid.UUID
	Type      TimelineEventType
	Lat       *float64
	Lon       *float64
	Message   string
	Meta      map[string]any
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e TimelineEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: timeline id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if !e.Type.Valid() {
		return fmt.Errorf("%w: invalid timeline type %q", ErrInvalidArgument, e.Type)
	}
	return nil
}

// GeofenceEventKind is enter or exit.
type GeofenceEventKind string

const (
	GeofenceEnter GeofenceEventKind = "enter"
	GeofenceExit  GeofenceEventKind = "exit"
)

// GeofenceEvent records a zone enter/exit for a courier/order.
type GeofenceEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	CourierID uuid.UUID
	OrderID   *uuid.UUID
	ZoneID    string
	Kind      GeofenceEventKind
	Lat       float64
	Lon       float64
	OccurredAt time.Time
	CreatedAt  time.Time
}

// ArrivalResult is the outcome of DetectArrival.
type ArrivalResult struct {
	Arrived          bool
	DistanceMeters   float64
	ThresholdMeters  float64
	CourierID        uuid.UUID
	OrderID          uuid.UUID
	TimelineEventID  *uuid.UUID
}
