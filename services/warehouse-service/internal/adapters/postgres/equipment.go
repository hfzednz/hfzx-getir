package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// EquipmentRepo persists equipment registry.
type EquipmentRepo struct{ DB *sql.DB }

var _ ports.EquipmentRepo = (*EquipmentRepo)(nil)

func (r *EquipmentRepo) Create(ctx context.Context, e domain.Equipment) error {
	if err := ensureWarehouse(ctx, r.DB, e.TenantID, e.WarehouseID); err != nil {
		return err
	}
	kind := e.Kind
	if kind == "" {
		kind = string(domain.EquipmentKindOther)
	}
	status := e.Status
	if status == "" {
		status = domain.EquipmentStatusOffline
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO equipment (
			id, tenant_id, warehouse_id, station_id, code, kind, status, name, serial_number, firmware,
			last_heartbeat, metadata, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		e.ID, e.TenantID, e.WarehouseID, nullUUID(e.StationID), e.Code, kind, string(status), e.Name, e.SerialNumber, e.Firmware,
		nullTime(e.LastHeartbeat), JSONMap(metaGetMap(e.Metadata)), e.CreatedAt, e.UpdatedAt, nullTime(e.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *EquipmentRepo) Update(ctx context.Context, e domain.Equipment) error {
	kind := e.Kind
	if kind == "" {
		kind = string(domain.EquipmentKindOther)
	}
	status := e.Status
	if status == "" {
		status = domain.EquipmentStatusOffline
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE equipment SET
			station_id=$1, code=$2, kind=$3, status=$4, name=$5, serial_number=$6, firmware=$7,
			last_heartbeat=$8, metadata=$9, updated_at=$10, deleted_at=$11
		WHERE id=$12 AND tenant_id=$13`,
		nullUUID(e.StationID), e.Code, kind, string(status), e.Name, e.SerialNumber, e.Firmware,
		nullTime(e.LastHeartbeat), JSONMap(metaGetMap(e.Metadata)), e.UpdatedAt, nullTime(e.DeletedAt),
		e.ID, e.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *EquipmentRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Equipment, error) {
	return r.scanEquipment(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, station_id, code, kind, status, name, serial_number, firmware,
			last_heartbeat, metadata, created_at, updated_at, deleted_at
		FROM equipment WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *EquipmentRepo) ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Equipment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, station_id, code, kind, status, name, serial_number, firmware,
			last_heartbeat, metadata, created_at, updated_at, deleted_at
		FROM equipment
		WHERE tenant_id=$1 AND warehouse_id=$2 AND deleted_at IS NULL
		ORDER BY code`, tenantID, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Equipment, 0)
	for rows.Next() {
		e, err := scanEquipmentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EquipmentRepo) scanEquipment(row scannable) (domain.Equipment, error) {
	e, err := scanEquipmentRow(row)
	if err != nil {
		return domain.Equipment{}, mapNotFound(err)
	}
	return e, nil
}

func scanEquipmentRow(row scannable) (domain.Equipment, error) {
	var e domain.Equipment
	var stationID uuid.NullUUID
	var status string
	var heartbeat, deleted sql.NullTime
	var meta JSONMap
	err := row.Scan(
		&e.ID, &e.TenantID, &e.WarehouseID, &stationID, &e.Code, &e.Kind, &status, &e.Name, &e.SerialNumber, &e.Firmware,
		&heartbeat, &meta, &e.CreatedAt, &e.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.Equipment{}, err
	}
	e.StationID = scanNullUUID(stationID)
	e.Status = domain.EquipmentStatus(status)
	e.LastHeartbeat = scanNullTime(heartbeat)
	e.DeletedAt = scanNullTime(deleted)
	e.Metadata = map[string]any(meta)
	e.CreatedAt = e.CreatedAt.UTC()
	e.UpdatedAt = e.UpdatedAt.UTC()
	return e, nil
}
