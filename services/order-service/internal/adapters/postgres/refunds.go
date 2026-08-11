package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// RefundRepo persists refund requests.
type RefundRepo struct{ DB *sql.DB }

var _ ports.RefundRepository = (*RefundRepo)(nil)

const refundColumns = `
	id, order_id, tenant_id, return_id, amount_minor, currency, method, status, reason,
	payment_refund_ref, actor_id, metadata, requested_at, completed_at, created_at, updated_at
`

func (r *RefundRepo) Create(ctx context.Context, ref domain.Refund) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO refunds (
			id, order_id, tenant_id, return_id, amount_minor, currency, method, status, reason,
			payment_refund_ref, actor_id, metadata, requested_at, completed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		ref.ID, ref.OrderID, ref.TenantID, nullUUID(ref.ReturnID), ref.AmountMinor, ref.Currency,
		string(ref.Method), string(ref.Status), ref.Reason, ref.PaymentRefundRef, nullUUID(ref.ActorID),
		JSONMap(ref.Metadata), ref.RequestedAt, nullTime(ref.CompletedAt), ref.CreatedAt, ref.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *RefundRepo) Update(ctx context.Context, ref domain.Refund) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE refunds SET
			return_id=$1, amount_minor=$2, currency=$3, method=$4, status=$5, reason=$6,
			payment_refund_ref=$7, actor_id=$8, metadata=$9, completed_at=$10, updated_at=$11
		WHERE id=$12 AND tenant_id=$13`,
		nullUUID(ref.ReturnID), ref.AmountMinor, ref.Currency, string(ref.Method), string(ref.Status),
		ref.Reason, ref.PaymentRefundRef, nullUUID(ref.ActorID), JSONMap(ref.Metadata),
		nullTime(ref.CompletedAt), ref.UpdatedAt, ref.ID, ref.TenantID,
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
	return nil
}

func (r *RefundRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Refund, error) {
	return scanRefund(r.DB.QueryRowContext(ctx, `
		SELECT `+refundColumns+` FROM refunds WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *RefundRepo) ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Refund, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+refundColumns+` FROM refunds
		WHERE tenant_id=$1 AND order_id=$2
		ORDER BY created_at DESC`, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Refund, 0)
	for rows.Next() {
		ref, err := scanRefund(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}

func scanRefund(row rowScanner) (domain.Refund, error) {
	var ref domain.Refund
	var method, status string
	var meta JSONMap
	var returnID, actor uuid.NullUUID
	var completed sql.NullTime
	err := row.Scan(
		&ref.ID, &ref.OrderID, &ref.TenantID, &returnID, &ref.AmountMinor, &ref.Currency, &method, &status,
		&ref.Reason, &ref.PaymentRefundRef, &actor, &meta, &ref.RequestedAt, &completed, &ref.CreatedAt, &ref.UpdatedAt,
	)
	if isNoRows(err) {
		return domain.Refund{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Refund{}, err
	}
	ref.Method = domain.RefundMethod(method)
	ref.Status = domain.RefundStatus(status)
	ref.ReturnID = scanNullUUID(returnID)
	ref.ActorID = scanNullUUID(actor)
	ref.Metadata = map[string]any(meta)
	if ref.Metadata == nil {
		ref.Metadata = map[string]any{}
	}
	ref.CompletedAt = scanNullTime(completed)
	ref.RequestedAt = ref.RequestedAt.UTC()
	ref.CreatedAt = ref.CreatedAt.UTC()
	ref.UpdatedAt = ref.UpdatedAt.UTC()
	return ref, nil
}
