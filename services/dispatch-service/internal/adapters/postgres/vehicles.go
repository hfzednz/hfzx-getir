package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// VehicleRepo persists fleet vehicles.
type VehicleRepo struct{ DB *sql.DB }

func (r *VehicleRepo) Upsert(ctx context.Context, v domain.Vehicle) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO vehicles (id, tenant_id, plate, type, capacity, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			plate=EXCLUDED.plate, type=EXCLUDED.type, capacity=EXCLUDED.capacity,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		v.ID, v.TenantID, v.Plate, string(v.Type), v.Capacity, v.Active, v.CreatedAt.UTC(), v.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *VehicleRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Vehicle, error) {
	v, err := scanVehicle(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, plate, type, capacity, active, created_at, updated_at
		FROM vehicles WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Vehicle{}, fmt.Errorf("%w: vehicle %s", domain.ErrNotFound, id)
	}
	return v, err
}

func (r *VehicleRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Vehicle, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicles WHERE tenant_id=$1`, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, plate, type, capacity, active, created_at, updated_at
		FROM vehicles WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Vehicle{}
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func scanVehicle(row scanner) (domain.Vehicle, error) {
	var v domain.Vehicle
	var typ string
	err := row.Scan(&v.ID, &v.TenantID, &v.Plate, &typ, &v.Capacity, &v.Active, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.Vehicle{}, err
	}
	v.Type = domain.VehicleType(typ)
	v.CreatedAt = v.CreatedAt.UTC()
	v.UpdatedAt = v.UpdatedAt.UTC()
	return v, nil
}

var _ ports.VehicleRepo = (*VehicleRepo)(nil)
