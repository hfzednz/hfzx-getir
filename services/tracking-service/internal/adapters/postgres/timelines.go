package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

// TimelineRepo persists delivery timeline and geofence events.
type TimelineRepo struct{ DB *sql.DB }

func (r *TimelineRepo) Append(ctx context.Context, e domain.TimelineEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO delivery_timelines (
			id, tenant_id, order_id, courier_id, type, lat, lon, message, meta, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.TenantID, e.OrderID, nullUUID(e.CourierID), string(e.Type),
		nullFloat(e.Lat), nullFloat(e.Lon), e.Message, JSONMap(e.Meta),
		e.OccurredAt.UTC(), e.CreatedAt.UTC())
	return err
}

func (r *TimelineRepo) ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID, limit int) ([]domain.TimelineEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, order_id, courier_id, type, lat, lon, message, meta, occurred_at, created_at
		FROM delivery_timelines
		WHERE tenant_id=$1 AND order_id=$2
		ORDER BY occurred_at ASC, created_at ASC
		LIMIT $3`, tenantID, orderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TimelineEvent{}
	for rows.Next() {
		var e domain.TimelineEvent
		var courier uuid.NullUUID
		var typ string
		var lat, lon sql.NullFloat64
		var meta JSONMap
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.OrderID, &courier, &typ, &lat, &lon, &e.Message, &meta,
			&e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.CourierID = scanNullUUID(courier)
		e.Type = domain.TimelineEventType(typ)
		e.Lat = scanNullFloat(lat)
		e.Lon = scanNullFloat(lon)
		e.Meta = map[string]any(meta)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *TimelineRepo) SaveGeofenceEvent(ctx context.Context, e domain.GeofenceEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO geofence_events (
			id, tenant_id, courier_id, order_id, zone_id, kind, lat, lon, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, e.TenantID, e.CourierID, nullUUID(e.OrderID), e.ZoneID, string(e.Kind),
		e.Lat, e.Lon, e.OccurredAt.UTC(), e.CreatedAt.UTC())
	return err
}

func (r *TimelineRepo) ListGeofenceEvents(ctx context.Context, tenantID uuid.UUID, courierID *uuid.UUID, limit int) ([]domain.GeofenceEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	var (
		rows *sql.Rows
		err  error
	)
	if courierID != nil {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, courier_id, order_id, zone_id, kind, lat, lon, occurred_at, created_at
			FROM geofence_events
			WHERE tenant_id=$1 AND courier_id=$2
			ORDER BY occurred_at ASC, created_at ASC
			LIMIT $3`, tenantID, *courierID, limit)
	} else {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, courier_id, order_id, zone_id, kind, lat, lon, occurred_at, created_at
			FROM geofence_events
			WHERE tenant_id=$1
			ORDER BY occurred_at ASC, created_at ASC
			LIMIT $2`, tenantID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GeofenceEvent{}
	for rows.Next() {
		var e domain.GeofenceEvent
		var order uuid.NullUUID
		var kind string
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.CourierID, &order, &e.ZoneID, &kind, &e.Lat, &e.Lon,
			&e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.OrderID = scanNullUUID(order)
		e.Kind = domain.GeofenceEventKind(kind)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

var _ ports.TimelineRepo = (*TimelineRepo)(nil)
