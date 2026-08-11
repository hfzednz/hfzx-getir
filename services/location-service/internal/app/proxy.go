package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// ProxyRouteInput proxies CreateRoute to routing-service.
type ProxyRouteInput struct {
	TenantID  uuid.UUID
	Origin    domain.LatLng
	Dest      domain.LatLng
	Waypoints []domain.LatLng
}

// ProxyRoute delegates route creation to RoutingClient.
func (d *Deps) ProxyRoute(ctx context.Context, in ProxyRouteInput) (ports.CreateRouteResult, error) {
	if in.TenantID == uuid.Nil {
		return ports.CreateRouteResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if err := in.Origin.Validate(); err != nil {
		return ports.CreateRouteResult{}, err
	}
	if err := in.Dest.Validate(); err != nil {
		return ports.CreateRouteResult{}, err
	}
	if d.Routing == nil {
		return ports.CreateRouteResult{}, fmt.Errorf("%w: routing client not configured", domain.ErrInvariant)
	}
	res, err := d.Routing.CreateRoute(ctx, ports.CreateRouteRequest{
		TenantID: in.TenantID, Origin: in.Origin, Dest: in.Dest, Waypoints: in.Waypoints,
	})
	if err != nil {
		return ports.CreateRouteResult{}, err
	}
	agg := d.newID()
	d.emit(ctx, in.TenantID, agg, domain.EventRouteCalculated, map[string]any{
		"routeId": res.RouteID, "distanceMeters": res.DistanceMeters,
	})
	return res, nil
}

// ProxyETAInput proxies ETA to routing-service.
type ProxyETAInput struct {
	TenantID uuid.UUID
	Origin   domain.LatLng
	Dest     domain.LatLng
}

// ProxyETA delegates ETA to RoutingClient.
func (d *Deps) ProxyETA(ctx context.Context, in ProxyETAInput) (ports.ETAResult, error) {
	if in.TenantID == uuid.Nil {
		return ports.ETAResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if err := in.Origin.Validate(); err != nil {
		return ports.ETAResult{}, err
	}
	if err := in.Dest.Validate(); err != nil {
		return ports.ETAResult{}, err
	}
	if d.Routing == nil {
		return ports.ETAResult{}, fmt.Errorf("%w: routing client not configured", domain.ErrInvariant)
	}
	res, err := d.Routing.ETA(ctx, ports.ETARequest{
		TenantID: in.TenantID, Origin: in.Origin, Dest: in.Dest,
	})
	if err != nil {
		return ports.ETAResult{}, err
	}
	agg := d.newID()
	d.emit(ctx, in.TenantID, agg, domain.EventETAUpdated, map[string]any{
		"durationSeconds": res.DurationSeconds, "distanceMeters": res.DistanceMeters,
	})
	return res, nil
}

// CheckZoneServiceabilityInput proxies serviceability to geofence-service.
type CheckZoneServiceabilityInput struct {
	TenantID uuid.UUID
	Lat      float64
	Lng      float64
}

// CheckZoneServiceability delegates to GeofenceClient.
func (d *Deps) CheckZoneServiceability(ctx context.Context, in CheckZoneServiceabilityInput) (domain.DeliveryFeasibility, error) {
	if in.TenantID == uuid.Nil {
		return domain.DeliveryFeasibility{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLng(in.Lat, in.Lng) {
		return domain.DeliveryFeasibility{}, fmt.Errorf("%w: lat/lng out of range", domain.ErrInvalidArgument)
	}
	if d.Geofence == nil {
		return domain.DeliveryFeasibility{}, fmt.Errorf("%w: geofence client not configured", domain.ErrInvariant)
	}
	return d.Geofence.CheckServiceability(ctx, in.TenantID, in.Lat, in.Lng)
}
