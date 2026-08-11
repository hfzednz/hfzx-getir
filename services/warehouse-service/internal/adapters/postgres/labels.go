package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

const (
	metaTrackingCode = "_trackingCode"
	metaBarcode      = "_barcode"
	metaFormat       = "_format"
	metaPrintIntent  = "_printIntent"
)

// LabelRepo persists shipping labels.
type LabelRepo struct{ DB *sql.DB }

var _ ports.LabelRepo = (*LabelRepo)(nil)

func (r *LabelRepo) Create(ctx context.Context, l domain.Label) error {
	if err := ensureWarehouse(ctx, r.DB, l.TenantID, l.WarehouseID); err != nil {
		return err
	}
	kind := l.Kind
	if kind == "" {
		kind = domain.LabelKindShipping
	}
	status := l.Status
	if status == "" {
		status = domain.LabelStatusDraft
	}
	code := l.Code
	if code == "" {
		code = l.TrackingCode
	}
	if code == "" {
		code = l.Barcode
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO labels (
			id, tenant_id, warehouse_id, fulfillment_id, pack_session_id, dispatch_unit_id,
			kind, status, code, payload, printer_id, printed_at, voided_at, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
		)`,
		l.ID, l.TenantID, l.WarehouseID, nullUUIDValue(l.FulfillmentID), nullUUIDValue(l.PackSessionID), nullUUID(l.DispatchUnitID),
		string(kind), string(status), code, labelPayload(l), nullUUID(l.PrinterID), nullTime(l.PrintedAt), nullTime(l.VoidedAt),
		labelMeta(l), l.CreatedAt, l.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *LabelRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Label, error) {
	var l domain.Label
	var fulfillmentID, packSessionID, dispatchUnitID, printerID uuid.NullUUID
	var kind, status string
	var payload, meta JSONMap
	var printed, voided sql.NullTime
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, fulfillment_id, pack_session_id, dispatch_unit_id,
			kind, status, code, payload, printer_id, printed_at, voided_at, metadata, created_at, updated_at
		FROM labels WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(
		&l.ID, &l.TenantID, &l.WarehouseID, &fulfillmentID, &packSessionID, &dispatchUnitID,
		&kind, &status, &l.Code, &payload, &printerID, &printed, &voided, &meta, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return domain.Label{}, mapNotFound(err)
	}
	l.FulfillmentID = scanUUIDOrNil(fulfillmentID)
	l.PackSessionID = scanUUIDOrNil(packSessionID)
	l.DispatchUnitID = scanNullUUID(dispatchUnitID)
	l.PrinterID = scanNullUUID(printerID)
	l.Kind = domain.LabelKind(kind)
	l.Status = domain.LabelStatus(status)
	l.PrintedAt = scanNullTime(printed)
	l.VoidedAt = scanNullTime(voided)
	l.CreatedAt = l.CreatedAt.UTC()
	l.UpdatedAt = l.UpdatedAt.UTC()
	l.Payload = map[string]any(payload)
	userMeta := map[string]any{}
	for k, v := range meta {
		switch k {
		case metaTrackingCode:
			if str, ok := v.(string); ok {
				l.TrackingCode = str
			}
		case metaBarcode:
			if str, ok := v.(string); ok {
				l.Barcode = str
			}
		case metaFormat:
			if str, ok := v.(string); ok {
				l.Format = str
			}
		case metaPrintIntent:
			if str, ok := v.(string); ok {
				l.PrintIntent = str
			}
		default:
			userMeta[k] = v
		}
	}
	l.Metadata = userMeta
	return l, nil
}

func labelPayload(l domain.Label) JSONMap {
	p := metaGetMap(l.Payload)
	if l.TrackingCode != "" {
		p["trackingCode"] = l.TrackingCode
	}
	if l.Barcode != "" {
		p["barcode"] = l.Barcode
	}
	if l.Format != "" {
		p["format"] = l.Format
	}
	if l.PrintIntent != "" {
		p["printIntent"] = l.PrintIntent
	}
	return JSONMap(p)
}

func labelMeta(l domain.Label) JSONMap {
	extra := map[string]any{}
	if l.TrackingCode != "" {
		extra[metaTrackingCode] = l.TrackingCode
	}
	if l.Barcode != "" {
		extra[metaBarcode] = l.Barcode
	}
	if l.Format != "" {
		extra[metaFormat] = l.Format
	}
	if l.PrintIntent != "" {
		extra[metaPrintIntent] = l.PrintIntent
	}
	return mergeMeta(l.Metadata, extra)
}
