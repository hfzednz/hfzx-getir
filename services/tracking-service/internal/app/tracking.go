package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

// IngestLocationInput is a courier GPS ping.
type IngestLocationInput struct {
	TenantID   uuid.UUID
	CourierID  uuid.UUID
	Lat        float64
	Lon        float64
	AccuracyM  float64
	HeadingDeg *float64
	SpeedMPS   *float64
	RecordedAt *time.Time
	OrderID    *uuid.UUID // optional: append timeline LocationUpdated
}

// GetLiveCourierInput fetches latest courier location.
type GetLiveCourierInput struct {
	TenantID  uuid.UUID
	CourierID uuid.UUID
}

// GetOrderTimelineInput lists timeline events for an opaque order id.
type GetOrderTimelineInput struct {
	TenantID uuid.UUID
	OrderID  uuid.UUID
	Limit    int
}

// AppendTimelineInput adds a custom/projected timeline event.
type AppendTimelineInput struct {
	TenantID  uuid.UUID
	OrderID   uuid.UUID
	CourierID *uuid.UUID
	Type      domain.TimelineEventType
	Lat       *float64
	Lon       *float64
	Message   string
	Meta      map[string]any
	OccurredAt *time.Time
}

// DetectArrivalInput checks distance to dropoff against threshold.
type DetectArrivalInput struct {
	TenantID    uuid.UUID
	CourierID   uuid.UUID
	OrderID     uuid.UUID
	DropoffLat  float64
	DropoffLon  float64
	ThresholdM  float64
}

// ListNearbyInput finds couriers within a radius of a point.
type ListNearbyInput struct {
	TenantID uuid.UUID
	Lat      float64
	Lon      float64
	RadiusM  float64
	Limit    int
}

// IngestLocation updates latest location, appends capped history, emits events.
func (d *Deps) IngestLocation(ctx context.Context, in IngestLocationInput) (domain.CourierLocation, error) {
	if in.TenantID == uuid.Nil || in.CourierID == uuid.Nil {
		return domain.CourierLocation{}, fmt.Errorf("%w: tenant_id and courier_id required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLon(in.Lat, in.Lon) {
		return domain.CourierLocation{}, fmt.Errorf("%w: invalid lat/lon", domain.ErrInvalidArgument)
	}
	now := d.now()
	recorded := now
	if in.RecordedAt != nil {
		recorded = in.RecordedAt.UTC()
	}

	update := domain.LocationUpdate{
		ID: d.newID(), TenantID: in.TenantID, CourierID: in.CourierID,
		Lat: in.Lat, Lon: in.Lon, AccuracyM: in.AccuracyM,
		HeadingDeg: in.HeadingDeg, SpeedMPS: in.SpeedMPS,
		RecordedAt: recorded, ReceivedAt: now,
	}
	if err := update.Validate(); err != nil {
		return domain.CourierLocation{}, err
	}

	loc := domain.CourierLocation{
		TenantID: in.TenantID, CourierID: in.CourierID,
		Lat: in.Lat, Lon: in.Lon, AccuracyM: in.AccuracyM,
		HeadingDeg: in.HeadingDeg, SpeedMPS: in.SpeedMPS,
		RecordedAt: recorded, UpdatedAt: now,
	}
	if err := d.Locations.UpsertLatest(ctx, loc); err != nil {
		return domain.CourierLocation{}, err
	}
	hist := domain.LocationHistoryEntry{
		ID: update.ID, TenantID: in.TenantID, CourierID: in.CourierID,
		Lat: in.Lat, Lon: in.Lon, AccuracyM: in.AccuracyM,
		HeadingDeg: in.HeadingDeg, SpeedMPS: in.SpeedMPS,
		RecordedAt: recorded, ReceivedAt: now,
	}
	if err := d.Locations.AppendHistory(ctx, hist, d.historyCap()); err != nil {
		return domain.CourierLocation{}, err
	}

	d.emit(ctx, in.TenantID, in.CourierID, domain.EventLocationUpdated, map[string]any{
		"courierId": in.CourierID.String(), "lat": in.Lat, "lon": in.Lon,
	})

	if in.OrderID != nil && *in.OrderID != uuid.Nil {
		lat, lon := in.Lat, in.Lon
		_, _ = d.AppendTimeline(ctx, AppendTimelineInput{
			TenantID: in.TenantID, OrderID: *in.OrderID, CourierID: &in.CourierID,
			Type: domain.TimelineLocationUpdated, Lat: &lat, Lon: &lon,
			Message: "location updated", OccurredAt: &recorded,
		})
	}

	if d.Geofence != nil {
		res, err := d.Geofence.Check(ctx, ports.GeofenceCheckRequest{
			TenantID: in.TenantID, CourierID: in.CourierID, OrderID: in.OrderID,
			Lat: in.Lat, Lon: in.Lon,
		})
		if err != nil {
			slog.Default().Warn("tracking.geofence.check",
				"err", err, "courierId", in.CourierID.String(), "tenantId", in.TenantID.String())
		} else {
			for _, hit := range res.Hits {
				evType := domain.EventGeofenceEnter
				if hit.Kind == domain.GeofenceExit {
					evType = domain.EventGeofenceExit
				}
				ge := domain.GeofenceEvent{
					ID: d.newID(), TenantID: in.TenantID, CourierID: in.CourierID,
					OrderID: in.OrderID, ZoneID: hit.ZoneID, Kind: hit.Kind,
					Lat: in.Lat, Lon: in.Lon, OccurredAt: now, CreatedAt: now,
				}
				_ = d.Timelines.SaveGeofenceEvent(ctx, ge)
				d.emit(ctx, in.TenantID, in.CourierID, evType, map[string]any{
					"zoneId": hit.ZoneID, "kind": string(hit.Kind),
				})
				if in.OrderID != nil {
					tlType := domain.TimelineGeofenceEnter
					if hit.Kind == domain.GeofenceExit {
						tlType = domain.TimelineGeofenceExit
					}
					lat, lon := in.Lat, in.Lon
					_, _ = d.AppendTimeline(ctx, AppendTimelineInput{
						TenantID: in.TenantID, OrderID: *in.OrderID, CourierID: &in.CourierID,
						Type: tlType, Lat: &lat, Lon: &lon, Message: string(hit.Kind) + " " + hit.ZoneID,
					})
				}
			}
		}
	}

	return loc, nil
}

// GetLiveCourier returns the latest courier location.
func (d *Deps) GetLiveCourier(ctx context.Context, in GetLiveCourierInput) (domain.CourierLocation, error) {
	return d.Locations.GetLatest(ctx, in.TenantID, in.CourierID)
}

// GetOrderTimeline returns timeline events for an order.
func (d *Deps) GetOrderTimeline(ctx context.Context, in GetOrderTimelineInput) ([]domain.TimelineEvent, error) {
	return d.Timelines.ListByOrder(ctx, in.TenantID, in.OrderID, in.Limit)
}

// AppendTimeline appends a timeline projection event.
func (d *Deps) AppendTimeline(ctx context.Context, in AppendTimelineInput) (domain.TimelineEvent, error) {
	if in.TenantID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.TimelineEvent{}, fmt.Errorf("%w: tenant_id and order_id required", domain.ErrInvalidArgument)
	}
	typ := in.Type
	if typ == "" {
		typ = domain.TimelineCustom
	}
	now := d.now()
	occurred := now
	if in.OccurredAt != nil {
		occurred = in.OccurredAt.UTC()
	}
	ev := domain.TimelineEvent{
		ID: d.newID(), TenantID: in.TenantID, OrderID: in.OrderID,
		CourierID: in.CourierID, Type: typ, Lat: in.Lat, Lon: in.Lon,
		Message: in.Message, Meta: in.Meta, OccurredAt: occurred, CreatedAt: now,
	}
	if err := ev.Validate(); err != nil {
		return domain.TimelineEvent{}, err
	}
	if err := d.Timelines.Append(ctx, ev); err != nil {
		return domain.TimelineEvent{}, err
	}
	return ev, nil
}

// DetectArrival marks arrival when courier is within threshold of dropoff.
func (d *Deps) DetectArrival(ctx context.Context, in DetectArrivalInput) (domain.ArrivalResult, error) {
	if in.TenantID == uuid.Nil || in.CourierID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.ArrivalResult{}, fmt.Errorf("%w: tenant_id, courier_id, order_id required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLon(in.DropoffLat, in.DropoffLon) {
		return domain.ArrivalResult{}, fmt.Errorf("%w: invalid dropoff lat/lon", domain.ErrInvalidArgument)
	}
	loc, err := d.Locations.GetLatest(ctx, in.TenantID, in.CourierID)
	if err != nil {
		return domain.ArrivalResult{}, err
	}
	thresh := in.ThresholdM
	if thresh <= 0 {
		thresh = d.arrivalThreshold()
	}
	dist := domain.HaversineMeters(loc.Lat, loc.Lon, in.DropoffLat, in.DropoffLon)
	res := domain.ArrivalResult{
		Arrived: dist <= thresh, DistanceMeters: dist, ThresholdMeters: thresh,
		CourierID: in.CourierID, OrderID: in.OrderID,
	}
	if !res.Arrived {
		return res, nil
	}
	lat, lon := loc.Lat, loc.Lon
	ev, err := d.AppendTimeline(ctx, AppendTimelineInput{
		TenantID: in.TenantID, OrderID: in.OrderID, CourierID: &in.CourierID,
		Type: domain.TimelineArrived, Lat: &lat, Lon: &lon,
		Message: "arrived at dropoff",
		Meta: map[string]any{"distanceMeters": dist, "thresholdMeters": thresh},
	})
	if err != nil {
		return domain.ArrivalResult{}, err
	}
	res.TimelineEventID = &ev.ID
	d.emit(ctx, in.TenantID, in.OrderID, domain.EventArrived, map[string]any{
		"courierId": in.CourierID.String(), "orderId": in.OrderID.String(),
		"distanceMeters": dist,
	})
	return res, nil
}

// ListNearby returns couriers within radiusM of the given point.
func (d *Deps) ListNearby(ctx context.Context, in ListNearbyInput) ([]domain.CourierLocation, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLon(in.Lat, in.Lon) {
		return nil, fmt.Errorf("%w: invalid lat/lon", domain.ErrInvalidArgument)
	}
	radius := in.RadiusM
	if radius <= 0 {
		radius = 1000
	}
	return d.Locations.ListNearby(ctx, in.TenantID, in.Lat, in.Lon, radius, in.Limit)
}
