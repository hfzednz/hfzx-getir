package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// DispatchRepo persists dispatch units.
type DispatchRepo struct{ DB *sql.DB }

var _ ports.DispatchRepo = (*DispatchRepo)(nil)

func (r *DispatchRepo) Create(ctx context.Context, u domain.DispatchUnit) error {
	if err := ensureWarehouse(ctx, r.DB, u.TenantID, u.WarehouseID); err != nil {
		return err
	}
	status := u.Status
	if status == "" {
		status = domain.DispatchStatusQueued
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO dispatch_units (
			id, tenant_id, warehouse_id, fulfillment_id, task_id, pack_session_id, station_id, label_id,
			package_code, tracking_code, courier_ref, status, verified_at, handed_off_at, failed_at,
			metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,$15,
			$16,$17,$18
		)`,
		u.ID, u.TenantID, u.WarehouseID, u.FulfillmentID, nullUUIDValue(u.TaskID), nullUUID(u.PackSessionID), nullUUID(u.StationID), nullUUID(u.LabelID),
		u.PackageCode, u.TrackingCode, u.CourierRef, string(status), nullTime(u.VerifiedAt), nullTime(u.HandedOffAt), nullTime(u.FailedAt),
		JSONMap(metaGetMap(u.Metadata)), u.CreatedAt, u.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *DispatchRepo) Update(ctx context.Context, u domain.DispatchUnit) error {
	status := u.Status
	if status == "" {
		status = domain.DispatchStatusQueued
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE dispatch_units SET
			warehouse_id=$1, fulfillment_id=$2, task_id=$3, pack_session_id=$4, station_id=$5, label_id=$6,
			package_code=$7, tracking_code=$8, courier_ref=$9, status=$10, verified_at=$11, handed_off_at=$12,
			failed_at=$13, metadata=$14, updated_at=$15
		WHERE id=$16 AND tenant_id=$17`,
		u.WarehouseID, u.FulfillmentID, nullUUIDValue(u.TaskID), nullUUID(u.PackSessionID), nullUUID(u.StationID), nullUUID(u.LabelID),
		u.PackageCode, u.TrackingCode, u.CourierRef, string(status), nullTime(u.VerifiedAt), nullTime(u.HandedOffAt),
		nullTime(u.FailedAt), JSONMap(metaGetMap(u.Metadata)), u.UpdatedAt,
		u.ID, u.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *DispatchRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.DispatchUnit, error) {
	return r.scanUnit(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, task_id, pack_session_id, station_id, label_id,
			package_code, tracking_code, courier_ref, status, verified_at, handed_off_at, failed_at,
			metadata, created_at, updated_at
		FROM dispatch_units WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *DispatchRepo) GetByFulfillmentID(ctx context.Context, tenantID, fulfillmentID uuid.UUID) (domain.DispatchUnit, error) {
	return r.scanUnit(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, task_id, pack_session_id, station_id, label_id,
			package_code, tracking_code, courier_ref, status, verified_at, handed_off_at, failed_at,
			metadata, created_at, updated_at
		FROM dispatch_units WHERE fulfillment_id=$1 AND tenant_id=$2`, fulfillmentID, tenantID))
}

func (r *DispatchRepo) ListQueued(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.DispatchUnit, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM dispatch_units
		WHERE tenant_id=$1 AND warehouse_id=$2 AND status IN ('queued','verified')`,
		tenantID, warehouseID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, task_id, pack_session_id, station_id, label_id,
			package_code, tracking_code, courier_ref, status, verified_at, handed_off_at, failed_at,
			metadata, created_at, updated_at
		FROM dispatch_units
		WHERE tenant_id=$1 AND warehouse_id=$2 AND status IN ('queued','verified')
		ORDER BY created_at ASC
		LIMIT $3 OFFSET $4`, tenantID, warehouseID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]domain.DispatchUnit, 0)
	for rows.Next() {
		u, err := scanDispatchRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *DispatchRepo) scanUnit(row scannable) (domain.DispatchUnit, error) {
	u, err := scanDispatchRow(row)
	if err != nil {
		return domain.DispatchUnit{}, mapNotFound(err)
	}
	return u, nil
}

func scanDispatchRow(row scannable) (domain.DispatchUnit, error) {
	var u domain.DispatchUnit
	var taskID, packSessionID, stationID, labelID uuid.NullUUID
	var status string
	var verified, handed, failed sql.NullTime
	var meta JSONMap
	err := row.Scan(
		&u.ID, &u.TenantID, &u.WarehouseID, &u.FulfillmentID, &taskID, &packSessionID, &stationID, &labelID,
		&u.PackageCode, &u.TrackingCode, &u.CourierRef, &status, &verified, &handed, &failed,
		&meta, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return domain.DispatchUnit{}, err
	}
	u.TaskID = scanUUIDOrNil(taskID)
	u.PackSessionID = scanNullUUID(packSessionID)
	u.StationID = scanNullUUID(stationID)
	u.LabelID = scanNullUUID(labelID)
	u.Status = domain.DispatchUnitStatus(status)
	u.VerifiedAt = scanNullTime(verified)
	u.HandedOffAt = scanNullTime(handed)
	u.FailedAt = scanNullTime(failed)
	u.Metadata = map[string]any(meta)
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()
	return u, nil
}
