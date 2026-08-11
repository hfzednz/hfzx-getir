package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

type BalanceRepo struct{ DB *sql.DB }

func (r *BalanceRepo) Create(ctx context.Context, b domain.StockBalance) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stock_balances (
			id, tenant_id, warehouse_id, variant_id, sku_code, location_id, on_hand, reserved, blocked,
			incoming, in_transit, version, safety_min, reorder_point, max_stock, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		b.ID, b.TenantID, b.WarehouseID, b.VariantID, b.SKUCode, nullUUID(b.LocationID), b.OnHand, b.Reserved, b.Blocked,
		b.Incoming, b.InTransit, b.Version, b.SafetyMin, b.ReorderPoint, nullInt64(b.MaxStock), JSONMap(b.Metadata), b.CreatedAt, b.UpdatedAt)
	return mapUniqueViolation(err)
}

func (r *BalanceRepo) Update(ctx context.Context, b domain.StockBalance) error {
	// Optimistic lock: require previous version (caller already bumped Version).
	prev := b.Version - 1
	if prev < 1 {
		prev = 1
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE stock_balances SET on_hand=$2, reserved=$3, blocked=$4, incoming=$5, in_transit=$6, version=$7,
			safety_min=$8, reorder_point=$9, max_stock=$10, sku_code=$11, metadata=$12, updated_at=$13
		WHERE id=$1 AND tenant_id=$14 AND version=$15`,
		b.ID, b.OnHand, b.Reserved, b.Blocked, b.Incoming, b.InTransit, b.Version,
		b.SafetyMin, b.ReorderPoint, nullInt64(b.MaxStock), b.SKUCode, JSONMap(b.Metadata), b.UpdatedAt, b.TenantID, prev)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// Distinguish not found vs conflict.
		var cur int64
		err := r.DB.QueryRowContext(ctx, `SELECT version FROM stock_balances WHERE id=$1 AND tenant_id=$2`, b.ID, b.TenantID).Scan(&cur)
		if err != nil {
			return mapNotFound(err)
		}
		return domain.ErrVersionConflict
	}
	return nil
}

func (r *BalanceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.StockBalance, error) {
	b, err := scanBalance(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, location_id, on_hand, reserved, blocked,
			incoming, in_transit, version, safety_min, reorder_point, max_stock, metadata, created_at, updated_at
		FROM stock_balances WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.StockBalance{}, mapNotFound(err)
	}
	return b, nil
}

func (r *BalanceRepo) GetByKey(ctx context.Context, key ports.BalanceKey) (domain.StockBalance, error) {
	b, err := scanBalance(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, location_id, on_hand, reserved, blocked,
			incoming, in_transit, version, safety_min, reorder_point, max_stock, metadata, created_at, updated_at
		FROM stock_balances
		WHERE tenant_id=$1 AND warehouse_id=$2 AND variant_id=$3
			AND COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::uuid) =
			    COALESCE($4::uuid, '00000000-0000-0000-0000-000000000000'::uuid)`,
		key.TenantID, key.WarehouseID, key.VariantID, nullUUID(key.LocationID)))
	if err != nil {
		return domain.StockBalance{}, mapNotFound(err)
	}
	return b, nil
}

func (r *BalanceRepo) ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockBalance, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stock_balances WHERE tenant_id=$1 AND warehouse_id=$2`, tenantID, warehouseID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, location_id, on_hand, reserved, blocked,
			incoming, in_transit, version, safety_min, reorder_point, max_stock, metadata, created_at, updated_at
		FROM stock_balances WHERE tenant_id=$1 AND warehouse_id=$2
		ORDER BY variant_id LIMIT $3 OFFSET $4`, tenantID, warehouseID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.StockBalance{}
	for rows.Next() {
		b, err := scanBalance(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}

func (r *BalanceRepo) ListByVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.StockBalance, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, location_id, on_hand, reserved, blocked,
			incoming, in_transit, version, safety_min, reorder_point, max_stock, metadata, created_at, updated_at
		FROM stock_balances WHERE tenant_id=$1 AND variant_id=$2`, tenantID, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StockBalance{}
	for rows.Next() {
		b, err := scanBalance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

type balanceScanner interface {
	Scan(dest ...any) error
}

func scanBalance(s balanceScanner) (domain.StockBalance, error) {
	var b domain.StockBalance
	var loc uuid.NullUUID
	var max sql.NullInt64
	var meta JSONMap
	err := s.Scan(&b.ID, &b.TenantID, &b.WarehouseID, &b.VariantID, &b.SKUCode, &loc, &b.OnHand, &b.Reserved, &b.Blocked,
		&b.Incoming, &b.InTransit, &b.Version, &b.SafetyMin, &b.ReorderPoint, &max, &meta, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.StockBalance{}, err
	}
	b.LocationID = scanNullUUID(loc)
	b.MaxStock = scanNullInt64(max)
	b.Metadata = map[string]any(meta)
	return b, nil
}

var _ ports.BalanceRepository = (*BalanceRepo)(nil)

type LotRepo struct{ DB *sql.DB }

func (r *LotRepo) Create(ctx context.Context, l domain.Lot) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stock_lots (
			id, tenant_id, balance_id, warehouse_id, variant_id, lot_code, qty, expiry_date, mfg_date,
			status, received_at, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::lot_status,$11,$12,$13,$14)`,
		l.ID, l.TenantID, l.BalanceID, l.WarehouseID, l.VariantID, l.LotCode, l.Qty, dateOnly(l.ExpiryDate), dateOnly(l.MfgDate),
		string(l.Status), nullTime(l.ReceivedAt), JSONMap(l.Metadata), l.CreatedAt, l.UpdatedAt)
	return mapUniqueViolation(err)
}

func (r *LotRepo) Update(ctx context.Context, l domain.Lot) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE stock_lots SET qty=$2, expiry_date=$3, mfg_date=$4, status=$5::lot_status, received_at=$6,
			metadata=$7, updated_at=$8
		WHERE id=$1 AND tenant_id=$9`,
		l.ID, l.Qty, dateOnly(l.ExpiryDate), dateOnly(l.MfgDate), string(l.Status), nullTime(l.ReceivedAt),
		JSONMap(l.Metadata), l.UpdatedAt, l.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LotRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Lot, error) {
	l, err := scanLot(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, balance_id, warehouse_id, variant_id, lot_code, qty, expiry_date, mfg_date,
			status::text, received_at, metadata, created_at, updated_at
		FROM stock_lots WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Lot{}, mapNotFound(err)
	}
	return l, nil
}

func (r *LotRepo) ListByBalance(ctx context.Context, balanceID uuid.UUID) ([]domain.Lot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, balance_id, warehouse_id, variant_id, lot_code, qty, expiry_date, mfg_date,
			status::text, received_at, metadata, created_at, updated_at
		FROM stock_lots WHERE balance_id=$1`, balanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectLots(rows)
}

func (r *LotRepo) ListByWarehouseVariant(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.Lot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, balance_id, warehouse_id, variant_id, lot_code, qty, expiry_date, mfg_date,
			status::text, received_at, metadata, created_at, updated_at
		FROM stock_lots WHERE tenant_id=$1 AND warehouse_id=$2 AND variant_id=$3`, tenantID, warehouseID, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectLots(rows)
}

func (r *LotRepo) ListNearExpiry(ctx context.Context, tenantID uuid.UUID, warehouseID *uuid.UUID, withinDays int, asOf time.Time) ([]domain.Lot, error) {
	deadline := asOf.AddDate(0, 0, withinDays)
	q := `
		SELECT id, tenant_id, balance_id, warehouse_id, variant_id, lot_code, qty, expiry_date, mfg_date,
			status::text, received_at, metadata, created_at, updated_at
		FROM stock_lots WHERE tenant_id=$1 AND expiry_date IS NOT NULL AND expiry_date <= $2`
	args := []any{tenantID, dateOnly(&deadline)}
	if warehouseID != nil {
		q += ` AND warehouse_id=$3`
		args = append(args, *warehouseID)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lots, err := collectLots(rows)
	if err != nil {
		return nil, err
	}
	domain.SortLotsFEFO(lots)
	return lots, nil
}

func collectLots(rows *sql.Rows) ([]domain.Lot, error) {
	out := []domain.Lot{}
	for rows.Next() {
		l, err := scanLot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type lotScanner interface {
	Scan(dest ...any) error
}

func scanLot(s lotScanner) (domain.Lot, error) {
	var l domain.Lot
	var status string
	var expiry, mfg, received sql.NullTime
	var meta JSONMap
	err := s.Scan(&l.ID, &l.TenantID, &l.BalanceID, &l.WarehouseID, &l.VariantID, &l.LotCode, &l.Qty, &expiry, &mfg,
		&status, &received, &meta, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return domain.Lot{}, err
	}
	l.ExpiryDate = scanDate(expiry)
	l.MfgDate = scanDate(mfg)
	l.Status = domain.LotStatus(status)
	l.ReceivedAt = scanNullTime(received)
	l.Metadata = map[string]any(meta)
	return l, nil
}

var _ ports.LotRepository = (*LotRepo)(nil)
