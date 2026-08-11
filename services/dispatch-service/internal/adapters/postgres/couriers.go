package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// CourierPool persists courier availability snapshots.
type CourierPool struct{ DB *sql.DB }

func (r *CourierPool) Upsert(ctx context.Context, c domain.CourierSnapshot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO courier_snapshots (
			id, tenant_id, courier_principal_id, available, lat, lng, current_load, max_capacity,
			rating, vehicle_type, on_shift, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, courier_principal_id) DO UPDATE SET
			id=EXCLUDED.id, available=EXCLUDED.available, lat=EXCLUDED.lat, lng=EXCLUDED.lng,
			current_load=EXCLUDED.current_load, max_capacity=EXCLUDED.max_capacity, rating=EXCLUDED.rating,
			vehicle_type=EXCLUDED.vehicle_type, on_shift=EXCLUDED.on_shift, updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.CourierPrincipalID, c.Available, c.Lat, c.Lng, c.CurrentLoad, c.MaxCapacity,
		c.Rating, string(c.VehicleType), c.OnShift, c.UpdatedAt.UTC())
	return err
}

func (r *CourierPool) Get(ctx context.Context, tenantID, courierPrincipalID uuid.UUID) (domain.CourierSnapshot, error) {
	c, err := scanCourier(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, courier_principal_id, available, lat, lng, current_load, max_capacity,
			rating, vehicle_type, on_shift, updated_at
		FROM courier_snapshots WHERE tenant_id=$1 AND courier_principal_id=$2`, tenantID, courierPrincipalID))
	if isNoRows(err) {
		return domain.CourierSnapshot{}, fmt.Errorf("%w: courier %s", domain.ErrNotFound, courierPrincipalID)
	}
	return c, err
}

func (r *CourierPool) ListAvailable(ctx context.Context, tenantID uuid.UUID) ([]domain.CourierSnapshot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, courier_principal_id, available, lat, lng, current_load, max_capacity,
			rating, vehicle_type, on_shift, updated_at
		FROM courier_snapshots WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CourierSnapshot{}
	for rows.Next() {
		c, err := scanCourier(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CourierPool) AdjustLoad(ctx context.Context, tenantID, courierPrincipalID uuid.UUID, delta int) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE courier_snapshots SET
			current_load = GREATEST(0, current_load + $3),
			updated_at = now()
		WHERE tenant_id=$1 AND courier_principal_id=$2`, tenantID, courierPrincipalID, delta)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: courier %s", domain.ErrNotFound, courierPrincipalID)
	}
	return nil
}

func scanCourier(row scanner) (domain.CourierSnapshot, error) {
	var c domain.CourierSnapshot
	var vehicle string
	err := row.Scan(
		&c.ID, &c.TenantID, &c.CourierPrincipalID, &c.Available, &c.Lat, &c.Lng,
		&c.CurrentLoad, &c.MaxCapacity, &c.Rating, &vehicle, &c.OnShift, &c.UpdatedAt)
	if err != nil {
		return domain.CourierSnapshot{}, err
	}
	c.VehicleType = domain.VehicleType(vehicle)
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

var _ ports.CourierPool = (*CourierPool)(nil)
