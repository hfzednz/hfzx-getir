package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// FulfillmentRepo persists fulfillment order projections.
type FulfillmentRepo struct{ DB *sql.DB }

var _ ports.FulfillmentRepo = (*FulfillmentRepo)(nil)

func (r *FulfillmentRepo) Create(ctx context.Context, o domain.FulfillmentOrder) error {
	if err := ensureWarehouse(ctx, r.DB, o.TenantID, o.WarehouseID); err != nil {
		return err
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	strategy := o.Strategy
	if strategy == "" {
		strategy = domain.PickStrategySingle
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO fulfillment_orders (
			id, tenant_id, order_id, external_order_id, warehouse_id, reservation_id,
			status, priority, strategy, vip, express, sla_deadline, courier_ref, metadata,
			created_at, updated_at, cancelled_at, failed_at, completed_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,$13,$14,
			$15,$16,$17,$18,$19
		)`,
		o.ID, o.TenantID, nullUUIDValue(o.OrderID), o.ExternalOrderID, o.WarehouseID, nullUUID(o.ReservationID),
		string(o.Status), o.Priority, string(strategy), o.VIP, o.Express, nullTime(o.SLADeadline), o.CourierRef, JSONMap(metaGetMap(o.Metadata)),
		o.CreatedAt, o.UpdatedAt, nullTime(o.CancelledAt), nullTime(o.FailedAt), nullTime(o.CompletedAt),
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := insertFulfillmentLines(ctx, tx, o.ID, o.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *FulfillmentRepo) Update(ctx context.Context, o domain.FulfillmentOrder) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	strategy := o.Strategy
	if strategy == "" {
		strategy = domain.PickStrategySingle
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE fulfillment_orders SET
			order_id=$1, external_order_id=$2, warehouse_id=$3, reservation_id=$4,
			status=$5, priority=$6, strategy=$7, vip=$8, express=$9, sla_deadline=$10,
			courier_ref=$11, metadata=$12, updated_at=$13, cancelled_at=$14, failed_at=$15, completed_at=$16
		WHERE id=$17 AND tenant_id=$18`,
		nullUUIDValue(o.OrderID), o.ExternalOrderID, o.WarehouseID, nullUUID(o.ReservationID),
		string(o.Status), o.Priority, string(strategy), o.VIP, o.Express, nullTime(o.SLADeadline),
		o.CourierRef, JSONMap(metaGetMap(o.Metadata)), o.UpdatedAt, nullTime(o.CancelledAt), nullTime(o.FailedAt), nullTime(o.CompletedAt),
		o.ID, o.TenantID,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM fulfillment_lines WHERE fulfillment_id=$1`, o.ID); err != nil {
		return err
	}
	if err := insertFulfillmentLines(ctx, tx, o.ID, o.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func insertFulfillmentLines(ctx context.Context, tx *sql.Tx, fulfillmentID uuid.UUID, lines []domain.FulfillmentLine) error {
	for _, l := range lines {
		qty := lineQty(l)
		barcodeExpected := l.BarcodeExpected
		if barcodeExpected == "" {
			barcodeExpected = l.Barcode
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO fulfillment_lines (
				id, fulfillment_id, variant_id, sku_code, qty, qty_picked, qty_packed,
				location_code, barcode_expected, barcode, expiry_required, sort_order, sequence,
				metadata, created_at, updated_at
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,
				$8,$9,$10,$11,$12,$13,
				$14,$15,$16
			)`,
			l.ID, fulfillmentID, l.VariantID, l.SKUCode, qty, int(l.QtyPicked), int(l.QtyPacked),
			l.LocationCode, barcodeExpected, l.Barcode, l.ExpiryRequired, l.SortOrder, l.Sequence,
			JSONMap(metaGetMap(l.Metadata)), l.CreatedAt, l.UpdatedAt,
		)
		if err != nil {
			return mapUniqueViolation(err)
		}
	}
	return nil
}

func (r *FulfillmentRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.FulfillmentOrder, error) {
	o, err := r.scanOrder(ctx, r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, order_id, external_order_id, warehouse_id, reservation_id,
			status, priority, strategy, vip, express, sla_deadline, courier_ref, metadata,
			created_at, updated_at, cancelled_at, failed_at, completed_at
		FROM fulfillment_orders WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	lines, err := r.loadLines(ctx, o.ID)
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	o.Lines = lines
	return o, nil
}

func (r *FulfillmentRepo) GetByExternalOrderID(ctx context.Context, tenantID uuid.UUID, externalOrderID string) (domain.FulfillmentOrder, error) {
	o, err := r.scanOrder(ctx, r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, order_id, external_order_id, warehouse_id, reservation_id,
			status, priority, strategy, vip, express, sla_deadline, courier_ref, metadata,
			created_at, updated_at, cancelled_at, failed_at, completed_at
		FROM fulfillment_orders WHERE tenant_id=$1 AND external_order_id=$2`, tenantID, externalOrderID))
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	lines, err := r.loadLines(ctx, o.ID)
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	o.Lines = lines
	return o, nil
}

func (r *FulfillmentRepo) List(ctx context.Context, f ports.FulfillmentFilter) ([]domain.FulfillmentOrder, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	where := []string{"tenant_id = $1"}
	args := []any{f.TenantID}
	n := 2
	if f.WarehouseID != nil {
		where = append(where, fmt.Sprintf("warehouse_id = $%d", n))
		args = append(args, *f.WarehouseID)
		n++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, string(*f.Status))
		n++
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM fulfillment_orders WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, order_id, external_order_id, warehouse_id, reservation_id,
			status, priority, strategy, vip, express, sla_deadline, courier_ref, metadata,
			created_at, updated_at, cancelled_at, failed_at, completed_at
		FROM fulfillment_orders
		WHERE `+whereSQL+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(n)+` OFFSET $`+fmt.Sprint(n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]domain.FulfillmentOrder, 0)
	for rows.Next() {
		o, err := scanFulfillmentRow(rows)
		if err != nil {
			return nil, 0, err
		}
		lines, err := r.loadLines(ctx, o.ID)
		if err != nil {
			return nil, 0, err
		}
		o.Lines = lines
		out = append(out, o)
	}
	return out, total, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func (r *FulfillmentRepo) scanOrder(ctx context.Context, row scannable) (domain.FulfillmentOrder, error) {
	_ = ctx
	o, err := scanFulfillmentRow(row)
	if err != nil {
		return domain.FulfillmentOrder{}, mapNotFound(err)
	}
	return o, nil
}

func scanFulfillmentRow(row scannable) (domain.FulfillmentOrder, error) {
	var o domain.FulfillmentOrder
	var orderID, reservationID uuid.NullUUID
	var status, strategy string
	var meta JSONMap
	var sla, cancelled, failed, completed sql.NullTime
	err := row.Scan(
		&o.ID, &o.TenantID, &orderID, &o.ExternalOrderID, &o.WarehouseID, &reservationID,
		&status, &o.Priority, &strategy, &o.VIP, &o.Express, &sla, &o.CourierRef, &meta,
		&o.CreatedAt, &o.UpdatedAt, &cancelled, &failed, &completed,
	)
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	o.OrderID = scanUUIDOrNil(orderID)
	o.ReservationID = scanNullUUID(reservationID)
	o.Status = domain.FulfillmentStatus(status)
	o.Strategy = domain.PickStrategy(strategy)
	o.Metadata = map[string]any(meta)
	o.SLADeadline = scanNullTime(sla)
	o.CancelledAt = scanNullTime(cancelled)
	o.FailedAt = scanNullTime(failed)
	o.CompletedAt = scanNullTime(completed)
	o.CreatedAt = o.CreatedAt.UTC()
	o.UpdatedAt = o.UpdatedAt.UTC()
	return o, nil
}

func (r *FulfillmentRepo) loadLines(ctx context.Context, fulfillmentID uuid.UUID) ([]domain.FulfillmentLine, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, fulfillment_id, variant_id, sku_code, qty, qty_picked, qty_packed,
			location_code, barcode_expected, barcode, expiry_required, sort_order, sequence,
			metadata, created_at, updated_at
		FROM fulfillment_lines WHERE fulfillment_id=$1 ORDER BY sort_order, sequence, created_at`, fulfillmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FulfillmentLine, 0)
	for rows.Next() {
		var l domain.FulfillmentLine
		var qty, qtyPicked, qtyPacked int
		var meta JSONMap
		if err := rows.Scan(
			&l.ID, &l.FulfillmentID, &l.VariantID, &l.SKUCode, &qty, &qtyPicked, &qtyPacked,
			&l.LocationCode, &l.BarcodeExpected, &l.Barcode, &l.ExpiryRequired, &l.SortOrder, &l.Sequence,
			&meta, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.Qty = qty
		l.QtyOrdered = int64(qty)
		l.QtyPicked = int64(qtyPicked)
		l.QtyPacked = int64(qtyPacked)
		l.Metadata = map[string]any(meta)
		l.CreatedAt = l.CreatedAt.UTC()
		l.UpdatedAt = l.UpdatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}
