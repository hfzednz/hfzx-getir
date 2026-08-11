package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// PackRepo persists pack sessions.
type PackRepo struct{ DB *sql.DB }

var _ ports.PackRepo = (*PackRepo)(nil)

func (r *PackRepo) Create(ctx context.Context, s domain.PackSession) error {
	if err := ensureWarehouse(ctx, r.DB, s.TenantID, s.WarehouseID); err != nil {
		return err
	}
	status := s.Status
	if status == "" {
		status = domain.PackSessionStatusQueued
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO pack_sessions (
			id, tenant_id, task_id, warehouse_id, station_id, fulfillment_id, packer_id, status,
			expected_weight_g, weight_tolerance, actual_weight_g, weight_g, length_mm, width_mm, height_mm,
			materials, cold_chain, fragile, hazard, sealed_at, labeled_at, label_id, label_payload,
			started_at, completed_at, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,
			$9,$10,$11,$12,$13,$14,$15,
			$16,$17,$18,$19,$20,$21,$22,$23,
			$24,$25,$26,$27,$28
		)`,
		s.ID, s.TenantID, s.TaskID, s.WarehouseID, nullUUID(s.StationID), nullUUIDValue(s.FulfillmentID), nullUUID(s.PackerID), string(status),
		s.ExpectedWeightG, s.WeightTolerance, nullInt64(s.ActualWeightG), nullInt(s.WeightG), nullInt(s.LengthMM), nullInt(s.WidthMM), nullInt(s.HeightMM),
		JSONRaw{V: s.Materials}, s.ColdChain, s.Fragile, s.Hazard, nullTime(s.SealedAt), nullTime(s.LabeledAt), nullUUID(s.LabelID), JSONMap(metaGetMap(s.LabelPayload)),
		nullTime(s.StartedAt), nullTime(s.CompletedAt), JSONMap(metaGetMap(s.Metadata)), s.CreatedAt, s.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *PackRepo) Update(ctx context.Context, s domain.PackSession) error {
	status := s.Status
	if status == "" {
		status = domain.PackSessionStatusQueued
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE pack_sessions SET
			warehouse_id=$1, station_id=$2, fulfillment_id=$3, packer_id=$4, status=$5,
			expected_weight_g=$6, weight_tolerance=$7, actual_weight_g=$8, weight_g=$9,
			length_mm=$10, width_mm=$11, height_mm=$12, materials=$13, cold_chain=$14, fragile=$15, hazard=$16,
			sealed_at=$17, labeled_at=$18, label_id=$19, label_payload=$20, started_at=$21, completed_at=$22,
			metadata=$23, updated_at=$24
		WHERE id=$25 AND tenant_id=$26`,
		s.WarehouseID, nullUUID(s.StationID), nullUUIDValue(s.FulfillmentID), nullUUID(s.PackerID), string(status),
		s.ExpectedWeightG, s.WeightTolerance, nullInt64(s.ActualWeightG), nullInt(s.WeightG),
		nullInt(s.LengthMM), nullInt(s.WidthMM), nullInt(s.HeightMM), JSONRaw{V: s.Materials}, s.ColdChain, s.Fragile, s.Hazard,
		nullTime(s.SealedAt), nullTime(s.LabeledAt), nullUUID(s.LabelID), JSONMap(metaGetMap(s.LabelPayload)),
		nullTime(s.StartedAt), nullTime(s.CompletedAt), JSONMap(metaGetMap(s.Metadata)), s.UpdatedAt,
		s.ID, s.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *PackRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.PackSession, error) {
	return r.scanPack(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, task_id, warehouse_id, station_id, fulfillment_id, packer_id, status,
			expected_weight_g, weight_tolerance, actual_weight_g, weight_g, length_mm, width_mm, height_mm,
			materials, cold_chain, fragile, hazard, sealed_at, labeled_at, label_id, label_payload,
			started_at, completed_at, metadata, created_at, updated_at
		FROM pack_sessions WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *PackRepo) GetByTaskID(ctx context.Context, tenantID, taskID uuid.UUID) (domain.PackSession, error) {
	return r.scanPack(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, task_id, warehouse_id, station_id, fulfillment_id, packer_id, status,
			expected_weight_g, weight_tolerance, actual_weight_g, weight_g, length_mm, width_mm, height_mm,
			materials, cold_chain, fragile, hazard, sealed_at, labeled_at, label_id, label_payload,
			started_at, completed_at, metadata, created_at, updated_at
		FROM pack_sessions WHERE task_id=$1 AND tenant_id=$2`, taskID, tenantID))
}

func (r *PackRepo) scanPack(row scannable) (domain.PackSession, error) {
	var s domain.PackSession
	var stationID, fulfillmentID, packerID, labelID uuid.NullUUID
	var status string
	var actualWeight sql.NullInt64
	var weightG, lengthMM, widthMM, heightMM sql.NullInt64
	var materials TextArray
	var sealed, labeled, started, completed sql.NullTime
	var labelPayload, meta JSONMap
	err := row.Scan(
		&s.ID, &s.TenantID, &s.TaskID, &s.WarehouseID, &stationID, &fulfillmentID, &packerID, &status,
		&s.ExpectedWeightG, &s.WeightTolerance, &actualWeight, &weightG, &lengthMM, &widthMM, &heightMM,
		&materials, &s.ColdChain, &s.Fragile, &s.Hazard, &sealed, &labeled, &labelID, &labelPayload,
		&started, &completed, &meta, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return domain.PackSession{}, mapNotFound(err)
	}
	s.StationID = scanNullUUID(stationID)
	s.FulfillmentID = scanUUIDOrNil(fulfillmentID)
	s.PackerID = scanNullUUID(packerID)
	s.LabelID = scanNullUUID(labelID)
	s.Status = domain.PackSessionStatus(status)
	s.ActualWeightG = scanNullInt64(actualWeight)
	s.WeightG = scanNullInt(weightG)
	s.LengthMM = scanNullInt(lengthMM)
	s.WidthMM = scanNullInt(widthMM)
	s.HeightMM = scanNullInt(heightMM)
	s.Materials = []string(materials)
	s.SealedAt = scanNullTime(sealed)
	s.LabeledAt = scanNullTime(labeled)
	s.StartedAt = scanNullTime(started)
	s.CompletedAt = scanNullTime(completed)
	s.LabelPayload = map[string]any(labelPayload)
	s.Metadata = map[string]any(meta)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}
