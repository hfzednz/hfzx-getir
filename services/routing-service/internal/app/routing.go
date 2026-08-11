package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app/ports"
	"github.com/nexora/routing-service/internal/domain"
)

// WaypointInput is a create/optimize waypoint payload.
type WaypointInput struct {
	Kind    domain.WaypointKind
	Lat     float64
	Lon     float64
	OrderID *uuid.UUID
	Label   string
}

// CreateRouteInput creates a draft multi-stop route.
type CreateRouteInput struct {
	TenantID    uuid.UUID
	DispatchID  *uuid.UUID
	CourierID   *uuid.UUID
	WarehouseID *uuid.UUID
	Waypoints   []WaypointInput
	SpeedMPS    float64
}

// OptimizeInput reorders stops with nearest-neighbor and recalculates ETA.
type OptimizeInput struct {
	TenantID uuid.UUID
	RouteID  uuid.UUID
}

// RecalculateETAInput recomputes ETA, optionally moving the origin.
type RecalculateETAInput struct {
	TenantID uuid.UUID
	RouteID  uuid.UUID
	// Optional courier/current position — replaces first waypoint coords when set.
	CurrentLat *float64
	CurrentLon *float64
	Reason     string
}

// UpdateTrafficHintInput upserts a regional traffic factor.
type UpdateTrafficHintInput struct {
	TenantID   uuid.UUID
	HintID     *uuid.UUID
	RegionKey  string
	Lat        float64
	Lon        float64
	RadiusM    float64
	Factor     float64
	ValidFrom  *time.Time
	ValidUntil *time.Time
}

// GetRouteInput fetches a route by id.
type GetRouteInput struct {
	TenantID uuid.UUID
	RouteID  uuid.UUID
}

// CreateRoute persists a draft route and computes initial ETA.
func (d *Deps) CreateRoute(ctx context.Context, in CreateRouteInput) (domain.Route, error) {
	if in.TenantID == uuid.Nil {
		return domain.Route{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if len(in.Waypoints) == 0 {
		return domain.Route{}, fmt.Errorf("%w: waypoints required", domain.ErrInvalidArgument)
	}

	now := d.now()
	speed := in.SpeedMPS
	if speed <= 0 {
		speed = domain.DefaultSpeedMPS
	}

	wps := make([]domain.Waypoint, 0, len(in.Waypoints))
	for i, w := range in.Waypoints {
		kind := w.Kind
		if kind == "" {
			if i == 0 {
				kind = domain.WaypointWarehouse
			} else {
				kind = domain.WaypointStop
			}
		}
		wp := domain.Waypoint{
			ID: d.newID(), Sequence: i, Kind: kind,
			Lat: w.Lat, Lon: w.Lon, OrderID: w.OrderID, Label: w.Label,
		}
		if err := wp.Validate(); err != nil {
			return domain.Route{}, err
		}
		wps = append(wps, wp)
	}

	route := domain.Route{
		ID: d.newID(), TenantID: in.TenantID,
		DispatchID: in.DispatchID, CourierID: in.CourierID, WarehouseID: in.WarehouseID,
		Status: domain.RouteStatusDraft, Waypoints: wps,
		TrafficFactor: 1, WeatherFactor: 1, SpeedMPS: speed,
		CreatedAt: now, UpdatedAt: now,
	}

	if err := d.applyETA(ctx, &route, "create"); err != nil {
		return domain.Route{}, err
	}
	if err := route.Validate(); err != nil {
		return domain.Route{}, err
	}
	if err := d.Routes.SaveRoute(ctx, route); err != nil {
		return domain.Route{}, err
	}
	d.emit(ctx, route.TenantID, route.ID, domain.EventRouteUpdated, map[string]any{
		"status": string(route.Status), "distanceMeters": route.DistanceMeters,
	})
	return route, nil
}

// Optimize reorders multi-stop waypoints (nearest-neighbor) and recalculates ETA.
func (d *Deps) Optimize(ctx context.Context, in OptimizeInput) (domain.Route, error) {
	route, err := d.Routes.GetRoute(ctx, in.TenantID, in.RouteID)
	if err != nil {
		return domain.Route{}, err
	}
	route.Waypoints = domain.OptimizeNearestNeighbor(route.Waypoints)
	route.Status = domain.RouteStatusOptimized
	route.UpdatedAt = d.now()
	if err := d.applyETA(ctx, &route, "optimize"); err != nil {
		return domain.Route{}, err
	}
	if err := d.Routes.SaveRoute(ctx, route); err != nil {
		return domain.Route{}, err
	}
	d.emit(ctx, route.TenantID, route.ID, domain.EventRouteUpdated, map[string]any{
		"status": string(route.Status), "optimized": true,
	})
	return route, nil
}

// RecalculateETA refreshes distance/ETA after movement or factor changes.
func (d *Deps) RecalculateETA(ctx context.Context, in RecalculateETAInput) (domain.Route, error) {
	route, err := d.Routes.GetRoute(ctx, in.TenantID, in.RouteID)
	if err != nil {
		return domain.Route{}, err
	}
	if in.CurrentLat != nil && in.CurrentLon != nil {
		if len(route.Waypoints) == 0 {
			return domain.Route{}, fmt.Errorf("%w: no waypoints", domain.ErrInvariant)
		}
		if !domain.ValidLatLon(*in.CurrentLat, *in.CurrentLon) {
			return domain.Route{}, fmt.Errorf("%w: invalid current lat/lon", domain.ErrInvalidArgument)
		}
		route.Waypoints[0].Lat = *in.CurrentLat
		route.Waypoints[0].Lon = *in.CurrentLon
		if route.Waypoints[0].Kind == domain.WaypointWarehouse {
			route.Waypoints[0].Kind = domain.WaypointCourier
		}
	}
	reason := in.Reason
	if reason == "" {
		reason = "recalculate"
	}
	route.UpdatedAt = d.now()
	if err := d.applyETA(ctx, &route, reason); err != nil {
		return domain.Route{}, err
	}
	if err := d.Routes.SaveRoute(ctx, route); err != nil {
		return domain.Route{}, err
	}
	d.emit(ctx, route.TenantID, route.ID, domain.EventETAUpdated, map[string]any{
		"durationSeconds": route.DurationSeconds, "distanceMeters": route.DistanceMeters,
		"trafficFactor": route.TrafficFactor, "weatherFactor": route.WeatherFactor,
	})
	return route, nil
}

// UpdateTrafficHint upserts a traffic hint used by RecalculateETA / applyETA.
func (d *Deps) UpdateTrafficHint(ctx context.Context, in UpdateTrafficHintInput) (domain.TrafficHint, error) {
	if in.TenantID == uuid.Nil {
		return domain.TrafficHint{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.Factor <= 0 {
		return domain.TrafficHint{}, fmt.Errorf("%w: factor must be > 0", domain.ErrInvalidArgument)
	}
	now := d.now()
	id := d.newID()
	if in.HintID != nil && *in.HintID != uuid.Nil {
		id = *in.HintID
	}
	from := now
	if in.ValidFrom != nil {
		from = in.ValidFrom.UTC()
	}
	until := now.Add(2 * time.Hour)
	if in.ValidUntil != nil {
		until = in.ValidUntil.UTC()
	}
	radius := in.RadiusM
	if radius <= 0 {
		radius = 2000
	}
	hint := domain.TrafficHint{
		ID: id, TenantID: in.TenantID, RegionKey: in.RegionKey,
		Lat: in.Lat, Lon: in.Lon, RadiusM: radius, Factor: in.Factor,
		ValidFrom: from, ValidUntil: until, CreatedAt: now, UpdatedAt: now,
	}
	if err := hint.Validate(); err != nil {
		return domain.TrafficHint{}, err
	}
	if err := d.Routes.UpsertTrafficHint(ctx, hint); err != nil {
		return domain.TrafficHint{}, err
	}
	return hint, nil
}

// GetRoute returns a route by id.
func (d *Deps) GetRoute(ctx context.Context, in GetRouteInput) (domain.Route, error) {
	return d.Routes.GetRoute(ctx, in.TenantID, in.RouteID)
}

func (d *Deps) applyETA(ctx context.Context, route *domain.Route, reason string) error {
	if len(route.Waypoints) == 0 {
		return fmt.Errorf("%w: no waypoints", domain.ErrInvariant)
	}
	now := d.now()
	speed := route.SpeedMPS
	if speed <= 0 {
		speed = domain.DefaultSpeedMPS
		route.SpeedMPS = speed
	}

	traffic := 1.0
	weather := 1.0
	origin := route.Waypoints[0]
	if d.Traffic != nil {
		f, err := d.Traffic.Factor(ctx, ports.TrafficFactorRequest{
			TenantID: route.TenantID, Lat: origin.Lat, Lon: origin.Lon, At: now,
		})
		if err != nil {
			return err
		}
		if f > 0 {
			traffic = f
		}
	}
	if d.Weather != nil {
		f, err := d.Weather.Factor(ctx, ports.WeatherFactorRequest{
			TenantID: route.TenantID, Lat: origin.Lat, Lon: origin.Lon, At: now,
		})
		if err != nil {
			return err
		}
		if f > 0 {
			weather = f
		}
	}
	// Local traffic hints override / multiply when covering the origin.
	if d.Routes != nil {
		hints, err := d.Routes.ListActiveTrafficHints(ctx, route.TenantID, now)
		if err != nil {
			return err
		}
		for _, h := range hints {
			if h.RadiusM <= 0 {
				continue
			}
			if domain.HaversineMeters(origin.Lat, origin.Lon, h.Lat, h.Lon) <= h.RadiusM {
				traffic *= h.Factor
			}
		}
	}
	route.TrafficFactor = traffic
	route.WeatherFactor = weather

	legs := make([]domain.RouteLeg, 0, max(0, len(route.Waypoints)-1))
	var totalDist, totalDur float64
	cum := 0.0
	for i := 0; i < len(route.Waypoints); i++ {
		route.Waypoints[i].Sequence = i
		if i == 0 {
			eta := now
			route.Waypoints[i].ETAAt = &eta
			continue
		}
		prev := route.Waypoints[i-1]
		cur := route.Waypoints[i]
		dist := domain.HaversineMeters(prev.Lat, prev.Lon, cur.Lat, cur.Lon)
		if d.Maps != nil {
			mx, err := d.Maps.DistanceMatrix(ctx, ports.DistanceMatrixRequest{
				Origins:      []ports.LatLon{{Lat: prev.Lat, Lon: prev.Lon}},
				Destinations: []ports.LatLon{{Lat: cur.Lat, Lon: cur.Lon}},
			})
			if err != nil {
				return err
			}
			if len(mx.DistancesMeters) > 0 && len(mx.DistancesMeters[0]) > 0 {
				dist = mx.DistancesMeters[0][0]
			}
		}
		dur := domain.ETASeconds(dist, speed, traffic, weather)
		legs = append(legs, domain.RouteLeg{
			ID: d.newID(), RouteID: route.ID,
			FromSequence: i - 1, ToSequence: i,
			DistanceMeters: dist, DurationSeconds: dur,
		})
		totalDist += dist
		totalDur += dur
		cum += dur
		eta := now.Add(time.Duration(cum * float64(time.Second)))
		route.Waypoints[i].ETAAt = &eta
	}
	route.Legs = legs
	route.DistanceMeters = totalDist
	route.DurationSeconds = totalDur
	if totalDur > 0 {
		eta := now.Add(time.Duration(totalDur * float64(time.Second)))
		route.ETAAt = &eta
	} else {
		route.ETAAt = &now
	}

	snap := domain.ETASnapshot{
		ID: d.newID(), TenantID: route.TenantID, RouteID: route.ID,
		DistanceMeters: totalDist, DurationSeconds: totalDur,
		ETAAt: *route.ETAAt, TrafficFactor: traffic, WeatherFactor: weather,
		SpeedMPS: speed, Reason: reason, CapturedAt: now,
	}
	if d.Routes != nil {
		_ = d.Routes.SaveETASnapshot(ctx, snap)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
