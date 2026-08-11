package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// OrderRepo persists the order aggregate in PostgreSQL.
type OrderRepo struct{ DB *sql.DB }

var _ ports.OrderRepository = (*OrderRepo)(nil)

const orderColumns = `
	id, tenant_id, customer_principal_id, status, type, currency,
	subtotal_minor, discount_minor, tax_minor, shipping_minor, tip_minor, total_minor,
	address_snapshot, notes, gift, priority, warehouse_ids, version, idempotency_key,
	scheduled_at, placed_at, payment_intent_ref, reservation_ref, courier_ref,
	parent_order_id, metadata, created_at, updated_at, cancelled_at, completed_at, archived_at
`

func (r *OrderRepo) Create(ctx context.Context, o domain.Order) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	gift := NullJSONMap{Map: o.Gift, Valid: o.Gift != nil}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			id, tenant_id, customer_principal_id, status, type, currency,
			subtotal_minor, discount_minor, tax_minor, shipping_minor, tip_minor, total_minor,
			address_snapshot, notes, gift, priority, warehouse_ids, version, idempotency_key,
			scheduled_at, placed_at, payment_intent_ref, reservation_ref, courier_ref,
			parent_order_id, metadata, created_at, updated_at, cancelled_at, completed_at, archived_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,
			$13,$14,$15,$16,$17,$18,$19,
			$20,$21,$22,$23,$24,
			$25,$26,$27,$28,$29,$30,$31
		)`,
		o.ID, o.TenantID, o.CustomerPrincipalID, string(o.Status), string(o.Type), o.Currency,
		o.SubtotalMinor, o.DiscountMinor, o.TaxMinor, o.ShippingMinor, o.TipMinor, o.TotalMinor,
		JSONMap(o.AddressSnapshot), o.Notes, gift, o.Priority, UUIDArray(o.WarehouseIDs), o.Version, o.IdempotencyKey,
		nullTime(o.ScheduledAt), nullTime(o.PlacedAt), o.PaymentIntentRef, o.ReservationRef, o.CourierRef,
		nullUUID(o.ParentOrderID), JSONMap(o.Metadata), o.CreatedAt, o.UpdatedAt,
		nullTime(o.CancelledAt), nullTime(o.CompletedAt), nullTime(o.ArchivedAt),
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := insertOrderLines(ctx, tx, o.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *OrderRepo) Update(ctx context.Context, o domain.Order) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	gift := NullJSONMap{Map: o.Gift, Valid: o.Gift != nil}
	expectedVersion := o.Version - 1
	if expectedVersion < 0 {
		expectedVersion = 0
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE orders SET
			customer_principal_id=$1, status=$2, type=$3, currency=$4,
			subtotal_minor=$5, discount_minor=$6, tax_minor=$7, shipping_minor=$8, tip_minor=$9, total_minor=$10,
			address_snapshot=$11, notes=$12, gift=$13, priority=$14, warehouse_ids=$15, version=$16,
			idempotency_key=$17, scheduled_at=$18, placed_at=$19,
			payment_intent_ref=$20, reservation_ref=$21, courier_ref=$22,
			parent_order_id=$23, metadata=$24, updated_at=$25,
			cancelled_at=$26, completed_at=$27, archived_at=$28
		WHERE id=$29 AND tenant_id=$30 AND version = $31`,
		o.CustomerPrincipalID, string(o.Status), string(o.Type), o.Currency,
		o.SubtotalMinor, o.DiscountMinor, o.TaxMinor, o.ShippingMinor, o.TipMinor, o.TotalMinor,
		JSONMap(o.AddressSnapshot), o.Notes, gift, o.Priority, UUIDArray(o.WarehouseIDs), o.Version,
		o.IdempotencyKey, nullTime(o.ScheduledAt), nullTime(o.PlacedAt),
		o.PaymentIntentRef, o.ReservationRef, o.CourierRef,
		nullUUID(o.ParentOrderID), JSONMap(o.Metadata), o.UpdatedAt,
		nullTime(o.CancelledAt), nullTime(o.CompletedAt), nullTime(o.ArchivedAt),
		o.ID, o.TenantID, expectedVersion,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var curVersion int64
		err := tx.QueryRowContext(ctx, `SELECT version FROM orders WHERE id=$1 AND tenant_id=$2`, o.ID, o.TenantID).Scan(&curVersion)
		if isNoRows(err) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		return domain.ErrVersionConflict
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM order_lines WHERE order_id=$1`, o.ID); err != nil {
		return err
	}
	if err := insertOrderLines(ctx, tx, o.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *OrderRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Order, error) {
	o, err := r.scanOrder(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+orderColumns+` FROM orders WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Order{}, err
	}
	lines, err := loadOrderLines(ctx, r.DB, o.ID)
	if err != nil {
		return domain.Order{}, err
	}
	o.Lines = lines
	return o, nil
}

func (r *OrderRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Order, error) {
	o, err := r.scanOrder(ctx, r.DB.QueryRowContext(ctx, `
		SELECT `+orderColumns+` FROM orders WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if err != nil {
		return domain.Order{}, err
	}
	lines, err := loadOrderLines(ctx, r.DB, o.ID)
	if err != nil {
		return domain.Order{}, err
	}
	o.Lines = lines
	return o, nil
}

func (r *OrderRepo) List(ctx context.Context, f ports.OrderFilter) ([]domain.Order, int, error) {
	where := []string{"tenant_id = $1"}
	args := []any{f.TenantID}
	argN := 2
	if f.CustomerID != nil {
		where = append(where, fmt.Sprintf("customer_principal_id = $%d", argN))
		args = append(args, *f.CustomerID)
		argN++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, string(*f.Status))
		argN++
	}
	if f.Query != "" {
		q := "%" + strings.ToLower(f.Query) + "%"
		where = append(where, fmt.Sprintf(
			`(LOWER(id::text) LIKE $%d OR LOWER(idempotency_key) LIKE $%d OR LOWER(status::text) LIKE $%d)`,
			argN, argN, argN))
		args = append(args, q)
		argN++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+orderColumns+` FROM orders WHERE `+whereSQL+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(argN)+` OFFSET $`+fmt.Sprint(argN+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Order, 0)
	for rows.Next() {
		o, err := scanOrderRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	for i := range out {
		lines, err := loadOrderLines(ctx, r.DB, out[i].ID)
		if err != nil {
			return nil, 0, err
		}
		out[i].Lines = lines
	}
	return out, total, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *OrderRepo) scanOrder(ctx context.Context, row rowScanner) (domain.Order, error) {
	_ = ctx
	o, err := scanOrderRow(row)
	if isNoRows(err) {
		return domain.Order{}, domain.ErrNotFound
	}
	return o, err
}

func scanOrderRow(row rowScanner) (domain.Order, error) {
	var o domain.Order
	var status, typ string
	var addr, meta JSONMap
	var gift NullJSONMap
	var warehouses UUIDArray
	var scheduled, placed, cancelled, completed, archived sql.NullTime
	var parent uuid.NullUUID
	err := row.Scan(
		&o.ID, &o.TenantID, &o.CustomerPrincipalID, &status, &typ, &o.Currency,
		&o.SubtotalMinor, &o.DiscountMinor, &o.TaxMinor, &o.ShippingMinor, &o.TipMinor, &o.TotalMinor,
		&addr, &o.Notes, &gift, &o.Priority, &warehouses, &o.Version, &o.IdempotencyKey,
		&scheduled, &placed, &o.PaymentIntentRef, &o.ReservationRef, &o.CourierRef,
		&parent, &meta, &o.CreatedAt, &o.UpdatedAt, &cancelled, &completed, &archived,
	)
	if err != nil {
		return domain.Order{}, err
	}
	o.Status = domain.OrderStatus(status)
	o.Type = domain.OrderType(typ)
	o.AddressSnapshot = map[string]any(addr)
	if o.AddressSnapshot == nil {
		o.AddressSnapshot = map[string]any{}
	}
	if gift.Valid {
		o.Gift = gift.Map
	}
	o.WarehouseIDs = []uuid.UUID(warehouses)
	o.Metadata = map[string]any(meta)
	if o.Metadata == nil {
		o.Metadata = map[string]any{}
	}
	o.ScheduledAt = scanNullTime(scheduled)
	o.PlacedAt = scanNullTime(placed)
	o.CancelledAt = scanNullTime(cancelled)
	o.CompletedAt = scanNullTime(completed)
	o.ArchivedAt = scanNullTime(archived)
	o.ParentOrderID = scanNullUUID(parent)
	o.CreatedAt = o.CreatedAt.UTC()
	o.UpdatedAt = o.UpdatedAt.UTC()
	return o, nil
}

func insertOrderLines(ctx context.Context, tx *sql.Tx, lines []domain.OrderLine) error {
	for _, l := range lines {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO order_lines (
				id, order_id, tenant_id, variant_id, sku_code, title_snapshot, qty,
				unit_price_minor, discounts_minor, tax_minor, line_total_minor,
				warehouse_id, sort_order, metadata, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			l.ID, l.OrderID, l.TenantID, l.VariantID, l.SKUCode, l.TitleSnapshot, l.Qty,
			l.UnitPriceMinor, l.DiscountsMinor, l.TaxMinor, l.LineTotalMinor,
			nullUUID(l.WarehouseID), l.SortOrder, JSONMap(l.Metadata), l.CreatedAt, l.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadOrderLines(ctx context.Context, db *sql.DB, orderID uuid.UUID) ([]domain.OrderLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, order_id, tenant_id, variant_id, sku_code, title_snapshot, qty,
			unit_price_minor, discounts_minor, tax_minor, line_total_minor,
			warehouse_id, sort_order, metadata, created_at, updated_at
		FROM order_lines WHERE order_id=$1 ORDER BY sort_order, created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.OrderLine, 0)
	for rows.Next() {
		var l domain.OrderLine
		var wh uuid.NullUUID
		var meta JSONMap
		if err := rows.Scan(
			&l.ID, &l.OrderID, &l.TenantID, &l.VariantID, &l.SKUCode, &l.TitleSnapshot, &l.Qty,
			&l.UnitPriceMinor, &l.DiscountsMinor, &l.TaxMinor, &l.LineTotalMinor,
			&wh, &l.SortOrder, &meta, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.WarehouseID = scanNullUUID(wh)
		l.Metadata = map[string]any(meta)
		if l.Metadata == nil {
			l.Metadata = map[string]any{}
		}
		l.CreatedAt = l.CreatedAt.UTC()
		l.UpdatedAt = l.UpdatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}
