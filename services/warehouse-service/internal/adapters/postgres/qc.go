package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

const (
	metaDefectCodes = "_defectCodes"
	metaCompletedAt = "_completedAt"
)

// QCRepo persists QC inspections.
type QCRepo struct{ DB *sql.DB }

var _ ports.QCRepo = (*QCRepo)(nil)

func (r *QCRepo) Create(ctx context.Context, i domain.QCInspection) error {
	if err := ensureWarehouse(ctx, r.DB, i.TenantID, i.WarehouseID); err != nil {
		return err
	}
	result := i.Result
	if result == "" {
		result = domain.QCResultPending
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO qc_inspections (
			id, tenant_id, warehouse_id, fulfillment_id, task_id, station_id, dispatch_unit_id,
			inspector_id, result, checklist, notes, defects, inspected_at, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,$10,$11,$12,$13,$14,$15,$16
		)`,
		i.ID, i.TenantID, i.WarehouseID, nullUUIDValue(i.FulfillmentID), nullUUID(i.TaskID), nullUUID(i.StationID), nullUUID(i.DispatchUnitID),
		nullUUID(i.InspectorID), string(result), JSONRaw{V: i.Checklist}, i.Notes, JSONRaw{V: i.Defects},
		nullTime(i.InspectedAt), qcMeta(i), i.CreatedAt, i.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *QCRepo) Update(ctx context.Context, i domain.QCInspection) error {
	result := i.Result
	if result == "" {
		result = domain.QCResultPending
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE qc_inspections SET
			warehouse_id=$1, fulfillment_id=$2, task_id=$3, station_id=$4, dispatch_unit_id=$5,
			inspector_id=$6, result=$7, checklist=$8, notes=$9, defects=$10, inspected_at=$11,
			metadata=$12, updated_at=$13
		WHERE id=$14 AND tenant_id=$15`,
		i.WarehouseID, nullUUIDValue(i.FulfillmentID), nullUUID(i.TaskID), nullUUID(i.StationID), nullUUID(i.DispatchUnitID),
		nullUUID(i.InspectorID), string(result), JSONRaw{V: i.Checklist}, i.Notes, JSONRaw{V: i.Defects}, nullTime(i.InspectedAt),
		qcMeta(i), i.UpdatedAt, i.ID, i.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *QCRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.QCInspection, error) {
	return r.scanQC(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, task_id, station_id, dispatch_unit_id,
			inspector_id, result, checklist, notes, defects, inspected_at, metadata, created_at, updated_at
		FROM qc_inspections WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *QCRepo) ListByFulfillment(ctx context.Context, tenantID, fulfillmentID uuid.UUID) ([]domain.QCInspection, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, task_id, station_id, dispatch_unit_id,
			inspector_id, result, checklist, notes, defects, inspected_at, metadata, created_at, updated_at
		FROM qc_inspections
		WHERE tenant_id=$1 AND fulfillment_id=$2
		ORDER BY created_at DESC`, tenantID, fulfillmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.QCInspection, 0)
	for rows.Next() {
		i, err := scanQCRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func qcMeta(i domain.QCInspection) JSONMap {
	extra := map[string]any{}
	if len(i.DefectCodes) > 0 {
		extra[metaDefectCodes] = i.DefectCodes
	}
	if i.CompletedAt != nil {
		extra[metaCompletedAt] = i.CompletedAt.UTC().Format(timeRFC3339Nano)
	}
	return mergeMeta(i.Metadata, extra)
}

const timeRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"

func (r *QCRepo) scanQC(row scannable) (domain.QCInspection, error) {
	i, err := scanQCRow(row)
	if err != nil {
		return domain.QCInspection{}, mapNotFound(err)
	}
	return i, nil
}

func scanQCRow(row scannable) (domain.QCInspection, error) {
	var i domain.QCInspection
	var fulfillmentID, taskID, stationID, dispatchUnitID, inspectorID uuid.NullUUID
	var result string
	var checklistRaw, defectsRaw []byte
	var inspected sql.NullTime
	var meta JSONMap
	err := row.Scan(
		&i.ID, &i.TenantID, &i.WarehouseID, &fulfillmentID, &taskID, &stationID, &dispatchUnitID,
		&inspectorID, &result, &checklistRaw, &i.Notes, &defectsRaw, &inspected, &meta, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return domain.QCInspection{}, err
	}
	i.FulfillmentID = scanUUIDOrNil(fulfillmentID)
	i.TaskID = scanNullUUID(taskID)
	i.StationID = scanNullUUID(stationID)
	i.DispatchUnitID = scanNullUUID(dispatchUnitID)
	i.InspectorID = scanNullUUID(inspectorID)
	i.Result = domain.QCResult(result)
	i.InspectedAt = scanNullTime(inspected)
	i.CreatedAt = i.CreatedAt.UTC()
	i.UpdatedAt = i.UpdatedAt.UTC()
	if len(checklistRaw) > 0 && string(checklistRaw) != "null" {
		_ = json.Unmarshal(checklistRaw, &i.Checklist)
	}
	if i.Checklist == nil {
		i.Checklist = []map[string]any{}
	}
	if len(defectsRaw) > 0 && string(defectsRaw) != "null" {
		_ = json.Unmarshal(defectsRaw, &i.Defects)
	}
	if i.Defects == nil {
		i.Defects = []map[string]any{}
	}
	userMeta := map[string]any{}
	for k, v := range meta {
		switch k {
		case metaDefectCodes:
			b, _ := json.Marshal(v)
			_ = json.Unmarshal(b, &i.DefectCodes)
		case metaCompletedAt:
			if str, ok := v.(string); ok {
				if t, err := time.Parse(time.RFC3339Nano, str); err == nil {
					tt := t.UTC()
					i.CompletedAt = &tt
				}
			}
		default:
			userMeta[k] = v
		}
	}
	i.Metadata = userMeta
	return i, nil
}
