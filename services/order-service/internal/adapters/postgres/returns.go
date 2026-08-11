package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// ReturnRepo persists return requests and lines.
type ReturnRepo struct{ DB *sql.DB }

var _ ports.ReturnRepository = (*ReturnRepo)(nil)

const returnColumns = `
	id, order_id, tenant_id, status, disposition, reason, notes, actor_id, refund_id,
	metadata, requested_at, decided_at, completed_at, created_at, updated_at
`

func (r *ReturnRepo) Create(ctx context.Context, ret domain.Return) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO returns (
			id, order_id, tenant_id, status, disposition, reason, notes, actor_id, refund_id,
			metadata, requested_at, decided_at, completed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		ret.ID, ret.OrderID, ret.TenantID, string(ret.Status), string(ret.Disposition),
		ret.Reason, ret.Notes, nullUUID(ret.ActorID), nullUUID(ret.RefundID),
		JSONMap(ret.Metadata), ret.RequestedAt, nullTime(ret.DecidedAt), nullTime(ret.CompletedAt),
		ret.CreatedAt, ret.UpdatedAt,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := insertReturnLines(ctx, tx, ret.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReturnRepo) Update(ctx context.Context, ret domain.Return) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE returns SET
			status=$1, disposition=$2, reason=$3, notes=$4, actor_id=$5, refund_id=$6,
			metadata=$7, decided_at=$8, completed_at=$9, updated_at=$10
		WHERE id=$11 AND tenant_id=$12`,
		string(ret.Status), string(ret.Disposition), ret.Reason, ret.Notes,
		nullUUID(ret.ActorID), nullUUID(ret.RefundID), JSONMap(ret.Metadata),
		nullTime(ret.DecidedAt), nullTime(ret.CompletedAt), ret.UpdatedAt,
		ret.ID, ret.TenantID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM return_lines WHERE return_id=$1`, ret.ID); err != nil {
		return err
	}
	if err := insertReturnLines(ctx, tx, ret.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReturnRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Return, error) {
	ret, err := scanReturn(r.DB.QueryRowContext(ctx, `
		SELECT `+returnColumns+` FROM returns WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Return{}, err
	}
	lines, err := loadReturnLines(ctx, r.DB, ret.ID)
	if err != nil {
		return domain.Return{}, err
	}
	ret.Lines = lines
	return ret, nil
}

func (r *ReturnRepo) ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Return, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+returnColumns+` FROM returns
		WHERE tenant_id=$1 AND order_id=$2
		ORDER BY created_at DESC`, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Return, 0)
	for rows.Next() {
		ret, err := scanReturn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ret)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		lines, err := loadReturnLines(ctx, r.DB, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Lines = lines
	}
	return out, nil
}

func scanReturn(row rowScanner) (domain.Return, error) {
	var ret domain.Return
	var status, disposition string
	var meta JSONMap
	var actor, refund uuid.NullUUID
	var decided, completed sql.NullTime
	err := row.Scan(
		&ret.ID, &ret.OrderID, &ret.TenantID, &status, &disposition, &ret.Reason, &ret.Notes,
		&actor, &refund, &meta, &ret.RequestedAt, &decided, &completed, &ret.CreatedAt, &ret.UpdatedAt,
	)
	if isNoRows(err) {
		return domain.Return{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Return{}, err
	}
	ret.Status = domain.ReturnStatus(status)
	ret.Disposition = domain.ReturnDisposition(disposition)
	ret.ActorID = scanNullUUID(actor)
	ret.RefundID = scanNullUUID(refund)
	ret.Metadata = map[string]any(meta)
	if ret.Metadata == nil {
		ret.Metadata = map[string]any{}
	}
	ret.DecidedAt = scanNullTime(decided)
	ret.CompletedAt = scanNullTime(completed)
	ret.RequestedAt = ret.RequestedAt.UTC()
	ret.CreatedAt = ret.CreatedAt.UTC()
	ret.UpdatedAt = ret.UpdatedAt.UTC()
	return ret, nil
}

func insertReturnLines(ctx context.Context, tx *sql.Tx, lines []domain.ReturnLine) error {
	for _, l := range lines {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO return_lines (
				id, return_id, order_line_id, variant_id, qty, disposition, reason, metadata, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			l.ID, l.ReturnID, l.OrderLineID, l.VariantID, l.Qty, string(l.Disposition),
			l.Reason, JSONMap(l.Metadata), l.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadReturnLines(ctx context.Context, db *sql.DB, returnID uuid.UUID) ([]domain.ReturnLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, return_id, order_line_id, variant_id, qty, disposition, reason, metadata, created_at
		FROM return_lines WHERE return_id=$1 ORDER BY created_at ASC`, returnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ReturnLine, 0)
	for rows.Next() {
		var l domain.ReturnLine
		var disposition string
		var meta JSONMap
		if err := rows.Scan(
			&l.ID, &l.ReturnID, &l.OrderLineID, &l.VariantID, &l.Qty, &disposition, &l.Reason, &meta, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.Disposition = domain.ReturnDisposition(disposition)
		l.Metadata = map[string]any(meta)
		if l.Metadata == nil {
			l.Metadata = map[string]any{}
		}
		l.CreatedAt = l.CreatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}
