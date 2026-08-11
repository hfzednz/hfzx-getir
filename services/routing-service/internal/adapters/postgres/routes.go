package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app/ports"
	"github.com/nexora/routing-service/internal/domain"
)

// RouteRepo persists routes, ETA snapshots, and traffic hints.
type RouteRepo struct{ DB *sql.DB }

func (r *RouteRepo) SaveRoute(ctx context.Context, route domain.Route) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO routes (
			id, tenant_id, dispatch_id, courier_id, warehouse_id, status,
			distance_meters, duration_seconds, eta_at, traffic_factor, weather_factor,
			speed_mps, waypoints, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (id) DO UPDATE SET
			dispatch_id=EXCLUDED.dispatch_id,
			courier_id=EXCLUDED.courier_id,
			warehouse_id=EXCLUDED.warehouse_id,
			status=EXCLUDED.status,
			distance_meters=EXCLUDED.distance_meters,
			duration_seconds=EXCLUDED.duration_seconds,
			eta_at=EXCLUDED.eta_at,
			traffic_factor=EXCLUDED.traffic_factor,
			weather_factor=EXCLUDED.weather_factor,
			speed_mps=EXCLUDED.speed_mps,
			waypoints=EXCLUDED.waypoints,
			updated_at=EXCLUDED.updated_at`,
		route.ID, route.TenantID, nullUUID(route.DispatchID), nullUUID(route.CourierID), nullUUID(route.WarehouseID),
		string(route.Status), route.DistanceMeters, route.DurationSeconds, nullTime(route.ETAAt),
		route.TrafficFactor, route.WeatherFactor, route.SpeedMPS, WaypointsJSON(route.Waypoints),
		route.CreatedAt.UTC(), route.UpdatedAt.UTC())
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM route_legs WHERE route_id=$1 AND tenant_id=$2`, route.ID, route.TenantID); err != nil {
		return err
	}
	for _, leg := range route.Legs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO route_legs (
				id, tenant_id, route_id, from_sequence, to_sequence, distance_meters, duration_seconds, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			leg.ID, route.TenantID, route.ID, leg.FromSequence, leg.ToSequence,
			leg.DistanceMeters, leg.DurationSeconds, route.UpdatedAt.UTC()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RouteRepo) GetRoute(ctx context.Context, tenantID, routeID uuid.UUID) (domain.Route, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, dispatch_id, courier_id, warehouse_id, status,
			distance_meters, duration_seconds, eta_at, traffic_factor, weather_factor,
			speed_mps, waypoints, created_at, updated_at
		FROM routes WHERE id=$1 AND tenant_id=$2`, routeID, tenantID)
	route, err := scanRoute(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Route{}, fmt.Errorf("%w: route", domain.ErrNotFound)
		}
		return domain.Route{}, err
	}
	legs, err := r.loadLegs(ctx, tenantID, routeID)
	if err != nil {
		return domain.Route{}, err
	}
	route.Legs = legs
	return route, nil
}

func (r *RouteRepo) ListRoutes(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.Route, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, dispatch_id, courier_id, warehouse_id, status,
			distance_meters, duration_seconds, eta_at, traffic_factor, weather_factor,
			speed_mps, waypoints, created_at, updated_at
		FROM routes WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Route{}
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		legs, err := r.loadLegs(ctx, tenantID, route.ID)
		if err != nil {
			return nil, err
		}
		route.Legs = legs
		out = append(out, route)
	}
	return out, rows.Err()
}

func (r *RouteRepo) loadLegs(ctx context.Context, tenantID, routeID uuid.UUID) ([]domain.RouteLeg, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, route_id, from_sequence, to_sequence, distance_meters, duration_seconds
		FROM route_legs WHERE tenant_id=$1 AND route_id=$2 ORDER BY from_sequence ASC`,
		tenantID, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RouteLeg{}
	for rows.Next() {
		var leg domain.RouteLeg
		if err := rows.Scan(&leg.ID, &leg.RouteID, &leg.FromSequence, &leg.ToSequence, &leg.DistanceMeters, &leg.DurationSeconds); err != nil {
			return nil, err
		}
		out = append(out, leg)
	}
	return out, rows.Err()
}

func (r *RouteRepo) SaveETASnapshot(ctx context.Context, s domain.ETASnapshot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO eta_snapshots (
			id, tenant_id, route_id, distance_meters, duration_seconds, eta_at,
			traffic_factor, weather_factor, speed_mps, reason, captured_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.TenantID, s.RouteID, s.DistanceMeters, s.DurationSeconds, s.ETAAt.UTC(),
		s.TrafficFactor, s.WeatherFactor, s.SpeedMPS, s.Reason, s.CapturedAt.UTC())
	return err
}

func (r *RouteRepo) ListETASnapshots(ctx context.Context, tenantID, routeID uuid.UUID, limit int) ([]domain.ETASnapshot, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, route_id, distance_meters, duration_seconds, eta_at,
			traffic_factor, weather_factor, speed_mps, reason, captured_at
		FROM eta_snapshots WHERE tenant_id=$1 AND route_id=$2
		ORDER BY captured_at DESC LIMIT $3`, tenantID, routeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ETASnapshot{}
	for rows.Next() {
		var s domain.ETASnapshot
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.RouteID, &s.DistanceMeters, &s.DurationSeconds, &s.ETAAt,
			&s.TrafficFactor, &s.WeatherFactor, &s.SpeedMPS, &s.Reason, &s.CapturedAt); err != nil {
			return nil, err
		}
		s.ETAAt = s.ETAAt.UTC()
		s.CapturedAt = s.CapturedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *RouteRepo) UpsertTrafficHint(ctx context.Context, h domain.TrafficHint) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO traffic_hints (
			id, tenant_id, region_key, lat, lon, radius_m, factor, valid_from, valid_until, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			region_key=EXCLUDED.region_key,
			lat=EXCLUDED.lat,
			lon=EXCLUDED.lon,
			radius_m=EXCLUDED.radius_m,
			factor=EXCLUDED.factor,
			valid_from=EXCLUDED.valid_from,
			valid_until=EXCLUDED.valid_until,
			updated_at=EXCLUDED.updated_at`,
		h.ID, h.TenantID, h.RegionKey, h.Lat, h.Lon, h.RadiusM, h.Factor,
		h.ValidFrom.UTC(), h.ValidUntil.UTC(), h.CreatedAt.UTC(), h.UpdatedAt.UTC())
	return err
}

func (r *RouteRepo) GetTrafficHint(ctx context.Context, tenantID, id uuid.UUID) (domain.TrafficHint, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, region_key, lat, lon, radius_m, factor, valid_from, valid_until, created_at, updated_at
		FROM traffic_hints WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var h domain.TrafficHint
	err := row.Scan(
		&h.ID, &h.TenantID, &h.RegionKey, &h.Lat, &h.Lon, &h.RadiusM, &h.Factor,
		&h.ValidFrom, &h.ValidUntil, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.TrafficHint{}, fmt.Errorf("%w: traffic hint", domain.ErrNotFound)
		}
		return domain.TrafficHint{}, err
	}
	h.ValidFrom = h.ValidFrom.UTC()
	h.ValidUntil = h.ValidUntil.UTC()
	h.CreatedAt = h.CreatedAt.UTC()
	h.UpdatedAt = h.UpdatedAt.UTC()
	return h, nil
}

func (r *RouteRepo) ListActiveTrafficHints(ctx context.Context, tenantID uuid.UUID, at time.Time) ([]domain.TrafficHint, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, region_key, lat, lon, radius_m, factor, valid_from, valid_until, created_at, updated_at
		FROM traffic_hints
		WHERE tenant_id=$1 AND valid_from <= $2 AND valid_until > $2`,
		tenantID, at.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TrafficHint{}
	for rows.Next() {
		var h domain.TrafficHint
		if err := rows.Scan(
			&h.ID, &h.TenantID, &h.RegionKey, &h.Lat, &h.Lon, &h.RadiusM, &h.Factor,
			&h.ValidFrom, &h.ValidUntil, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		h.ValidFrom = h.ValidFrom.UTC()
		h.ValidUntil = h.ValidUntil.UTC()
		h.CreatedAt = h.CreatedAt.UTC()
		h.UpdatedAt = h.UpdatedAt.UTC()
		out = append(out, h)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanRoute(row scannable) (domain.Route, error) {
	var route domain.Route
	var dispatch, courier, warehouse uuid.NullUUID
	var eta sql.NullTime
	var status string
	var waypoints WaypointsJSON
	err := row.Scan(
		&route.ID, &route.TenantID, &dispatch, &courier, &warehouse, &status,
		&route.DistanceMeters, &route.DurationSeconds, &eta, &route.TrafficFactor, &route.WeatherFactor,
		&route.SpeedMPS, &waypoints, &route.CreatedAt, &route.UpdatedAt)
	if err != nil {
		return domain.Route{}, err
	}
	route.DispatchID = scanNullUUID(dispatch)
	route.CourierID = scanNullUUID(courier)
	route.WarehouseID = scanNullUUID(warehouse)
	route.Status = domain.RouteStatus(status)
	route.ETAAt = scanNullTime(eta)
	route.Waypoints = []domain.Waypoint(waypoints)
	route.CreatedAt = route.CreatedAt.UTC()
	route.UpdatedAt = route.UpdatedAt.UTC()
	return route, nil
}

var _ ports.RouteRepo = (*RouteRepo)(nil)
