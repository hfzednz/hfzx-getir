package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

type TransferRepo struct{ DB *sql.DB }

func (r *TransferRepo) Create(ctx context.Context, t domain.Transfer) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transfers (
			id, tenant_id, code, from_warehouse_id, to_warehouse_id, from_location_id, to_location_id, status,
			requested_by, approved_by, reason, metadata, created_at, updated_at, approved_at, shipped_at, completed_at, cancelled_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::transfer_status,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		t.ID, t.TenantID, t.Code, t.FromWarehouseID, t.ToWarehouseID, nullUUID(t.FromLocationID), nullUUID(t.ToLocationID), string(t.Status),
		nullUUID(t.RequestedBy), nullUUID(t.ApprovedBy), t.Reason, JSONMap(t.Metadata), t.CreatedAt, t.UpdatedAt,
		nullTime(t.ApprovedAt), nullTime(t.ShippedAt), nullTime(t.CompletedAt), nullTime(t.CancelledAt)); err != nil {
		return err
	}
	if err := replaceTransferLines(ctx, tx, t.ID, t.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *TransferRepo) Update(ctx context.Context, t domain.Transfer) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE transfers SET code=$2, from_warehouse_id=$3, to_warehouse_id=$4, from_location_id=$5, to_location_id=$6,
			status=$7::transfer_status, requested_by=$8, approved_by=$9, reason=$10, metadata=$11, updated_at=$12,
			approved_at=$13, shipped_at=$14, completed_at=$15, cancelled_at=$16
		WHERE id=$1 AND tenant_id=$17`,
		t.ID, t.Code, t.FromWarehouseID, t.ToWarehouseID, nullUUID(t.FromLocationID), nullUUID(t.ToLocationID),
		string(t.Status), nullUUID(t.RequestedBy), nullUUID(t.ApprovedBy), t.Reason, JSONMap(t.Metadata), t.UpdatedAt,
		nullTime(t.ApprovedAt), nullTime(t.ShippedAt), nullTime(t.CompletedAt), nullTime(t.CancelledAt), t.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if err := replaceTransferLines(ctx, tx, t.ID, t.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *TransferRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Transfer, error) {
	t, err := scanTransfer(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, from_warehouse_id, to_warehouse_id, from_location_id, to_location_id, status::text,
			requested_by, approved_by, reason, metadata, created_at, updated_at, approved_at, shipped_at, completed_at, cancelled_at
		FROM transfers WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Transfer{}, mapNotFound(err)
	}
	lines, err := listTransferLines(ctx, r.DB, id)
	if err != nil {
		return domain.Transfer{}, err
	}
	t.Lines = lines
	return t, nil
}

func (r *TransferRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Transfer, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM transfers WHERE tenant_id=$1`, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, from_warehouse_id, to_warehouse_id, from_location_id, to_location_id, status::text,
			requested_by, approved_by, reason, metadata, created_at, updated_at, approved_at, shipped_at, completed_at, cancelled_at
		FROM transfers WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Transfer{}
	for rows.Next() {
		t, err := scanTransfer(rows)
		if err != nil {
			return nil, 0, err
		}
		lines, err := listTransferLines(ctx, r.DB, t.ID)
		if err != nil {
			return nil, 0, err
		}
		t.Lines = lines
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func replaceTransferLines(ctx context.Context, tx *sql.Tx, transferID uuid.UUID, lines []domain.TransferLine) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM transfer_lines WHERE transfer_id=$1`, transferID); err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transfer_lines (id, transfer_id, variant_id, sku_code, lot_id, qty_requested, qty_shipped, qty_received, metadata, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			line.ID, transferID, line.VariantID, line.SKUCode, nullUUID(line.LotID), line.QtyRequested, line.QtyShipped, line.QtyReceived,
			JSONMap(line.Metadata), line.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func listTransferLines(ctx context.Context, db *sql.DB, transferID uuid.UUID) ([]domain.TransferLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, transfer_id, variant_id, sku_code, lot_id, qty_requested, qty_shipped, qty_received, metadata, created_at
		FROM transfer_lines WHERE transfer_id=$1 ORDER BY created_at`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TransferLine{}
	for rows.Next() {
		var line domain.TransferLine
		var lot uuid.NullUUID
		var meta JSONMap
		if err := rows.Scan(&line.ID, &line.TransferID, &line.VariantID, &line.SKUCode, &lot, &line.QtyRequested, &line.QtyShipped, &line.QtyReceived, &meta, &line.CreatedAt); err != nil {
			return nil, err
		}
		line.LotID = scanNullUUID(lot)
		line.Metadata = map[string]any(meta)
		out = append(out, line)
	}
	return out, rows.Err()
}

type transferScanner interface {
	Scan(dest ...any) error
}

func scanTransfer(s transferScanner) (domain.Transfer, error) {
	var t domain.Transfer
	var fromLoc, toLoc, requested, approved uuid.NullUUID
	var status string
	var meta JSONMap
	var approvedAt, shippedAt, completedAt, cancelledAt sql.NullTime
	err := s.Scan(&t.ID, &t.TenantID, &t.Code, &t.FromWarehouseID, &t.ToWarehouseID, &fromLoc, &toLoc, &status,
		&requested, &approved, &t.Reason, &meta, &t.CreatedAt, &t.UpdatedAt, &approvedAt, &shippedAt, &completedAt, &cancelledAt)
	if err != nil {
		return domain.Transfer{}, err
	}
	t.FromLocationID = scanNullUUID(fromLoc)
	t.ToLocationID = scanNullUUID(toLoc)
	t.Status = domain.TransferStatus(status)
	t.RequestedBy = scanNullUUID(requested)
	t.ApprovedBy = scanNullUUID(approved)
	t.Metadata = map[string]any(meta)
	t.ApprovedAt = scanNullTime(approvedAt)
	t.ShippedAt = scanNullTime(shippedAt)
	t.CompletedAt = scanNullTime(completedAt)
	t.CancelledAt = scanNullTime(cancelledAt)
	return t, nil
}

var _ ports.TransferRepository = (*TransferRepo)(nil)

type CountRepo struct{ DB *sql.DB }

func (r *CountRepo) Create(ctx context.Context, s domain.CountSession) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO count_sessions (
			id, tenant_id, warehouse_id, location_id, type, status, started_by, submitted_by, approved_by, notes,
			metadata, created_at, updated_at, started_at, submitted_at, approved_at, cancelled_at
		) VALUES ($1,$2,$3,$4,$5::count_session_type,$6::count_session_status,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		s.ID, s.TenantID, s.WarehouseID, nullUUID(s.LocationID), string(s.Type), string(s.Status),
		nullUUID(s.StartedBy), nullUUID(s.SubmittedBy), nullUUID(s.ApprovedBy), s.Notes,
		JSONMap(s.Metadata), s.CreatedAt, s.UpdatedAt, nullTime(s.StartedAt), nullTime(s.SubmittedAt), nullTime(s.ApprovedAt), nullTime(s.CancelledAt)); err != nil {
		return err
	}
	if err := replaceCountLines(ctx, tx, s.ID, s.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CountRepo) Update(ctx context.Context, s domain.CountSession) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		UPDATE count_sessions SET location_id=$2, type=$3::count_session_type, status=$4::count_session_status,
			started_by=$5, submitted_by=$6, approved_by=$7, notes=$8, metadata=$9, updated_at=$10,
			started_at=$11, submitted_at=$12, approved_at=$13, cancelled_at=$14
		WHERE id=$1 AND tenant_id=$15`,
		s.ID, nullUUID(s.LocationID), string(s.Type), string(s.Status),
		nullUUID(s.StartedBy), nullUUID(s.SubmittedBy), nullUUID(s.ApprovedBy), s.Notes, JSONMap(s.Metadata), s.UpdatedAt,
		nullTime(s.StartedAt), nullTime(s.SubmittedAt), nullTime(s.ApprovedAt), nullTime(s.CancelledAt), s.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if err := replaceCountLines(ctx, tx, s.ID, s.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CountRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.CountSession, error) {
	s, err := scanCount(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, location_id, type::text, status::text, started_by, submitted_by, approved_by, notes,
			metadata, created_at, updated_at, started_at, submitted_at, approved_at, cancelled_at
		FROM count_sessions WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.CountSession{}, mapNotFound(err)
	}
	lines, err := listCountLines(ctx, r.DB, id)
	if err != nil {
		return domain.CountSession{}, err
	}
	s.Lines = lines
	return s, nil
}

func (r *CountRepo) List(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.CountSession, int, error) {
	if limit <= 0 {
		limit = 50
	}
	args := []any{tenantID}
	where := `tenant_id=$1`
	if warehouseID != uuid.Nil {
		args = append(args, warehouseID)
		where += ` AND warehouse_id=$2`
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM count_sessions WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, location_id, type::text, status::text, started_by, submitted_by, approved_by, notes,
			metadata, created_at, updated_at, started_at, submitted_at, approved_at, cancelled_at
		FROM count_sessions WHERE `+where+` ORDER BY created_at DESC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.CountSession{}
	for rows.Next() {
		s, err := scanCount(rows)
		if err != nil {
			return nil, 0, err
		}
		lines, err := listCountLines(ctx, r.DB, s.ID)
		if err != nil {
			return nil, 0, err
		}
		s.Lines = lines
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func replaceCountLines(ctx context.Context, tx *sql.Tx, sessionID uuid.UUID, lines []domain.CountLine) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM count_lines WHERE session_id=$1`, sessionID); err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO count_lines (
				id, session_id, variant_id, sku_code, location_id, lot_id, system_qty, counted_qty, variance, approved, notes, metadata, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			line.ID, sessionID, line.VariantID, line.SKUCode, nullUUID(line.LocationID), nullUUID(line.LotID), line.SystemQty,
			nullInt64(line.CountedQty), nullInt64(line.Variance), nullBool(line.Approved), line.Notes, JSONMap(line.Metadata), line.CreatedAt, line.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

func listCountLines(ctx context.Context, db *sql.DB, sessionID uuid.UUID) ([]domain.CountLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, session_id, variant_id, sku_code, location_id, lot_id, system_qty, counted_qty, variance, approved, notes, metadata, created_at, updated_at
		FROM count_lines WHERE session_id=$1 ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CountLine{}
	for rows.Next() {
		var line domain.CountLine
		var loc, lot uuid.NullUUID
		var counted, variance sql.NullInt64
		var approved sql.NullBool
		var meta JSONMap
		if err := rows.Scan(&line.ID, &line.SessionID, &line.VariantID, &line.SKUCode, &loc, &lot, &line.SystemQty, &counted, &variance, &approved, &line.Notes, &meta, &line.CreatedAt, &line.UpdatedAt); err != nil {
			return nil, err
		}
		line.LocationID = scanNullUUID(loc)
		line.LotID = scanNullUUID(lot)
		line.CountedQty = scanNullInt64(counted)
		line.Variance = scanNullInt64(variance)
		line.Approved = scanNullBool(approved)
		line.Metadata = map[string]any(meta)
		out = append(out, line)
	}
	return out, rows.Err()
}

type countScanner interface {
	Scan(dest ...any) error
}

func scanCount(s countScanner) (domain.CountSession, error) {
	var sess domain.CountSession
	var loc, startedBy, submittedBy, approvedBy uuid.NullUUID
	var typ, status string
	var meta JSONMap
	var started, submitted, approved, cancelled sql.NullTime
	err := s.Scan(&sess.ID, &sess.TenantID, &sess.WarehouseID, &loc, &typ, &status, &startedBy, &submittedBy, &approvedBy, &sess.Notes,
		&meta, &sess.CreatedAt, &sess.UpdatedAt, &started, &submitted, &approved, &cancelled)
	if err != nil {
		return domain.CountSession{}, err
	}
	sess.LocationID = scanNullUUID(loc)
	sess.Type = domain.CountSessionType(typ)
	sess.Status = domain.CountSessionStatus(status)
	sess.StartedBy = scanNullUUID(startedBy)
	sess.SubmittedBy = scanNullUUID(submittedBy)
	sess.ApprovedBy = scanNullUUID(approvedBy)
	sess.Metadata = map[string]any(meta)
	sess.StartedAt = scanNullTime(started)
	sess.SubmittedAt = scanNullTime(submitted)
	sess.ApprovedAt = scanNullTime(approved)
	sess.CancelledAt = scanNullTime(cancelled)
	return sess, nil
}

var _ ports.CountRepository = (*CountRepo)(nil)
