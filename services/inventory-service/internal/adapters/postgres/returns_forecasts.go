package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

type ReturnRepo struct{ DB *sql.DB }

func (r *ReturnRepo) Create(ctx context.Context, ret domain.InventoryReturn) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO inventory_returns (
			id, tenant_id, warehouse_id, source, disposition, status, external_ref, actor_id, reason,
			metadata, created_at, updated_at, received_at, disposed_at
		) VALUES ($1,$2,$3,$4::return_source,$5::return_disposition,$6::return_status,$7,$8,$9,$10,$11,$12,$13,$14)`,
		ret.ID, ret.TenantID, ret.WarehouseID, string(ret.Source), string(ret.Disposition), string(ret.Status),
		ret.ExternalRef, nullUUID(ret.ActorID), ret.Reason, JSONMap(ret.Metadata), ret.CreatedAt, ret.UpdatedAt,
		nullTime(ret.ReceivedAt), nullTime(ret.DisposedAt)); err != nil {
		return err
	}
	if err := replaceReturnLines(ctx, tx, ret.ID, ret.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReturnRepo) Update(ctx context.Context, ret domain.InventoryReturn) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE inventory_returns SET source=$2::return_source, disposition=$3::return_disposition, status=$4::return_status,
			external_ref=$5, actor_id=$6, reason=$7, metadata=$8, updated_at=$9, received_at=$10, disposed_at=$11
		WHERE id=$1 AND tenant_id=$12`,
		ret.ID, string(ret.Source), string(ret.Disposition), string(ret.Status),
		ret.ExternalRef, nullUUID(ret.ActorID), ret.Reason, JSONMap(ret.Metadata), ret.UpdatedAt,
		nullTime(ret.ReceivedAt), nullTime(ret.DisposedAt), ret.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if err := replaceReturnLines(ctx, tx, ret.ID, ret.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReturnRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.InventoryReturn, error) {
	ret, err := scanReturn(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, source::text, disposition::text, status::text, external_ref, actor_id, reason,
			metadata, created_at, updated_at, received_at, disposed_at
		FROM inventory_returns WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.InventoryReturn{}, mapNotFound(err)
	}
	lines, err := listReturnLines(ctx, r.DB, id)
	if err != nil {
		return domain.InventoryReturn{}, err
	}
	ret.Lines = lines
	return ret, nil
}

func replaceReturnLines(ctx context.Context, tx *sql.Tx, returnID uuid.UUID, lines []domain.ReturnLine) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM inventory_return_lines WHERE return_id=$1`, returnID); err != nil {
		return err
	}
	for _, line := range lines {
		var disp any
		if line.Disposition != nil {
			disp = string(*line.Disposition)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inventory_return_lines (
				id, return_id, variant_id, sku_code, lot_id, location_id, qty, disposition, condition_notes, metadata, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::return_disposition,$9,$10,$11)`,
			line.ID, returnID, line.VariantID, line.SKUCode, nullUUID(line.LotID), nullUUID(line.LocationID), line.Qty,
			disp, line.ConditionNotes, JSONMap(line.Metadata), line.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func listReturnLines(ctx context.Context, db *sql.DB, returnID uuid.UUID) ([]domain.ReturnLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, return_id, variant_id, sku_code, lot_id, location_id, qty, disposition::text, condition_notes, metadata, created_at
		FROM inventory_return_lines WHERE return_id=$1 ORDER BY created_at`, returnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ReturnLine{}
	for rows.Next() {
		var line domain.ReturnLine
		var lot, loc uuid.NullUUID
		var disp sql.NullString
		var meta JSONMap
		if err := rows.Scan(&line.ID, &line.ReturnID, &line.VariantID, &line.SKUCode, &lot, &loc, &line.Qty, &disp, &line.ConditionNotes, &meta, &line.CreatedAt); err != nil {
			return nil, err
		}
		line.LotID = scanNullUUID(lot)
		line.LocationID = scanNullUUID(loc)
		if disp.Valid && disp.String != "" {
			d := domain.ReturnDisposition(disp.String)
			line.Disposition = &d
		}
		line.Metadata = map[string]any(meta)
		out = append(out, line)
	}
	return out, rows.Err()
}

type returnScanner interface {
	Scan(dest ...any) error
}

func scanReturn(s returnScanner) (domain.InventoryReturn, error) {
	var ret domain.InventoryReturn
	var source, disposition, status string
	var actor uuid.NullUUID
	var meta JSONMap
	var received, disposed sql.NullTime
	err := s.Scan(&ret.ID, &ret.TenantID, &ret.WarehouseID, &source, &disposition, &status, &ret.ExternalRef, &actor, &ret.Reason,
		&meta, &ret.CreatedAt, &ret.UpdatedAt, &received, &disposed)
	if err != nil {
		return domain.InventoryReturn{}, err
	}
	ret.Source = domain.ReturnSource(source)
	ret.Disposition = domain.ReturnDisposition(disposition)
	ret.Status = domain.ReturnStatus(status)
	ret.ActorID = scanNullUUID(actor)
	ret.Metadata = map[string]any(meta)
	ret.ReceivedAt = scanNullTime(received)
	ret.DisposedAt = scanNullTime(disposed)
	return ret, nil
}

var _ ports.ReturnRepository = (*ReturnRepo)(nil)

type ForecastRepo struct{ DB *sql.DB }

func (r *ForecastRepo) Upsert(ctx context.Context, f domain.StockForecast) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stock_forecasts (
			id, tenant_id, warehouse_id, variant_id, sku_code, horizon_start, horizon_end, predicted_demand,
			predicted_atp, confidence, model_id, model_version, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (warehouse_id, variant_id, horizon_start, horizon_end, model_id) DO UPDATE SET
			sku_code=EXCLUDED.sku_code, predicted_demand=EXCLUDED.predicted_demand, predicted_atp=EXCLUDED.predicted_atp,
			confidence=EXCLUDED.confidence, model_version=EXCLUDED.model_version, metadata=EXCLUDED.metadata,
			updated_at=EXCLUDED.updated_at, id=stock_forecasts.id`,
		f.ID, f.TenantID, f.WarehouseID, f.VariantID, f.SKUCode, dateOnly(&f.HorizonStart), dateOnly(&f.HorizonEnd), f.PredictedDemand,
		nullFloat64(f.PredictedATP), nullFloat64(f.Confidence), f.ModelID, f.ModelVersion, JSONMap(f.Metadata), f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *ForecastRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.StockForecast, error) {
	f, err := scanForecast(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, horizon_start, horizon_end, predicted_demand,
			predicted_atp, confidence, model_id, model_version, metadata, created_at, updated_at
		FROM stock_forecasts WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.StockForecast{}, mapNotFound(err)
	}
	return f, nil
}

func (r *ForecastRepo) List(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.StockForecast, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, variant_id, sku_code, horizon_start, horizon_end, predicted_demand,
			predicted_atp, confidence, model_id, model_version, metadata, created_at, updated_at
		FROM stock_forecasts WHERE tenant_id=$1 AND warehouse_id=$2 AND variant_id=$3
		ORDER BY horizon_start`, tenantID, warehouseID, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StockForecast{}
	for rows.Next() {
		f, err := scanForecast(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type forecastScanner interface {
	Scan(dest ...any) error
}

func scanForecast(s forecastScanner) (domain.StockForecast, error) {
	var f domain.StockForecast
	var atp, conf sql.NullFloat64
	var meta JSONMap
	err := s.Scan(&f.ID, &f.TenantID, &f.WarehouseID, &f.VariantID, &f.SKUCode, &f.HorizonStart, &f.HorizonEnd, &f.PredictedDemand,
		&atp, &conf, &f.ModelID, &f.ModelVersion, &meta, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return domain.StockForecast{}, err
	}
	f.PredictedATP = scanNullFloat64(atp)
	f.Confidence = scanNullFloat64(conf)
	f.Metadata = map[string]any(meta)
	f.HorizonStart = f.HorizonStart.UTC()
	f.HorizonEnd = f.HorizonEnd.UTC()
	return f, nil
}

var _ ports.ForecastRepository = (*ForecastRepo)(nil)
