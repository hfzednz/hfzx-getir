package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

// LocationRepo persists latest courier locations and capped history.
type LocationRepo struct{ DB *sql.DB }

func (r *LocationRepo) UpsertLatest(ctx context.Context, loc domain.CourierLocation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO courier_locations (
			tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, courier_id) DO UPDATE SET
			lat=EXCLUDED.lat,
			lon=EXCLUDED.lon,
			accuracy_m=EXCLUDED.accuracy_m,
			heading_deg=EXCLUDED.heading_deg,
			speed_mps=EXCLUDED.speed_mps,
			recorded_at=EXCLUDED.recorded_at,
			updated_at=EXCLUDED.updated_at`,
		loc.TenantID, loc.CourierID, loc.Lat, loc.Lon, loc.AccuracyM,
		nullFloat(loc.HeadingDeg), nullFloat(loc.SpeedMPS),
		loc.RecordedAt.UTC(), loc.UpdatedAt.UTC())
	return err
}

func (r *LocationRepo) GetLatest(ctx context.Context, tenantID, courierID uuid.UUID) (domain.CourierLocation, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, updated_at
		FROM courier_locations WHERE tenant_id=$1 AND courier_id=$2`, tenantID, courierID)
	var loc domain.CourierLocation
	var heading, speed sql.NullFloat64
	err := row.Scan(
		&loc.TenantID, &loc.CourierID, &loc.Lat, &loc.Lon, &loc.AccuracyM,
		&heading, &speed, &loc.RecordedAt, &loc.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.CourierLocation{}, fmt.Errorf("%w: courier location", domain.ErrNotFound)
		}
		return domain.CourierLocation{}, err
	}
	loc.HeadingDeg = scanNullFloat(heading)
	loc.SpeedMPS = scanNullFloat(speed)
	loc.RecordedAt = loc.RecordedAt.UTC()
	loc.UpdatedAt = loc.UpdatedAt.UTC()
	return loc, nil
}

func (r *LocationRepo) AppendHistory(ctx context.Context, entry domain.LocationHistoryEntry, cap int) error {
	if cap <= 0 {
		cap = domain.DefaultHistoryCap
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO location_history (
			id, tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, received_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		entry.ID, entry.TenantID, entry.CourierID, entry.Lat, entry.Lon, entry.AccuracyM,
		nullFloat(entry.HeadingDeg), nullFloat(entry.SpeedMPS),
		entry.RecordedAt.UTC(), entry.ReceivedAt.UTC())
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM location_history
		WHERE id IN (
			SELECT id FROM location_history
			WHERE tenant_id=$1 AND courier_id=$2
			ORDER BY recorded_at DESC, received_at DESC
			OFFSET $3
		)`, entry.TenantID, entry.CourierID, cap)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *LocationRepo) ListHistory(ctx context.Context, tenantID, courierID uuid.UUID, limit int) ([]domain.LocationHistoryEntry, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if limit <= 0 {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, received_at
			FROM location_history
			WHERE tenant_id=$1 AND courier_id=$2
			ORDER BY recorded_at ASC, received_at ASC`, tenantID, courierID)
	} else {
		rows, err = r.DB.QueryContext(ctx, `
			SELECT id, tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, received_at
			FROM (
				SELECT id, tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, received_at
				FROM location_history
				WHERE tenant_id=$1 AND courier_id=$2
				ORDER BY recorded_at DESC, received_at DESC
				LIMIT $3
			) recent
			ORDER BY recorded_at ASC, received_at ASC`, tenantID, courierID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LocationHistoryEntry{}
	for rows.Next() {
		var e domain.LocationHistoryEntry
		var heading, speed sql.NullFloat64
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.CourierID, &e.Lat, &e.Lon, &e.AccuracyM,
			&heading, &speed, &e.RecordedAt, &e.ReceivedAt); err != nil {
			return nil, err
		}
		e.HeadingDeg = scanNullFloat(heading)
		e.SpeedMPS = scanNullFloat(speed)
		e.RecordedAt = e.RecordedAt.UTC()
		e.ReceivedAt = e.ReceivedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *LocationRepo) ListNearby(ctx context.Context, tenantID uuid.UUID, lat, lon, radiusM float64, limit int) ([]domain.CourierLocation, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, courier_id, lat, lon, accuracy_m, heading_deg, speed_mps, recorded_at, updated_at
		FROM courier_locations WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CourierLocation{}
	for rows.Next() {
		var loc domain.CourierLocation
		var heading, speed sql.NullFloat64
		if err := rows.Scan(
			&loc.TenantID, &loc.CourierID, &loc.Lat, &loc.Lon, &loc.AccuracyM,
			&heading, &speed, &loc.RecordedAt, &loc.UpdatedAt); err != nil {
			return nil, err
		}
		if domain.HaversineMeters(lat, lon, loc.Lat, loc.Lon) > radiusM {
			continue
		}
		loc.HeadingDeg = scanNullFloat(heading)
		loc.SpeedMPS = scanNullFloat(speed)
		loc.RecordedAt = loc.RecordedAt.UTC()
		loc.UpdatedAt = loc.UpdatedAt.UTC()
		out = append(out, loc)
		if len(out) >= limit {
			break
		}
	}
	return out, rows.Err()
}

var _ ports.LocationRepo = (*LocationRepo)(nil)
