package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// FulfillmentRepo persists split fulfillment units.
type FulfillmentRepo struct{ DB *sql.DB }

var _ ports.FulfillmentRepository = (*FulfillmentRepo)(nil)

const fulfillmentColumns = `
	id, order_id, tenant_id, warehouse_id, status, reservation_id, fulfillment_ref,
	priority, line_ids, metadata, created_at, updated_at, cancelled_at, completed_at
`

func (r *FulfillmentRepo) Create(ctx context.Context, f domain.Fulfillment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO fulfillments (
			id, order_id, tenant_id, warehouse_id, status, reservation_id, fulfillment_ref,
			priority, line_ids, metadata, created_at, updated_at, cancelled_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		f.ID, f.OrderID, f.TenantID, f.WarehouseID, string(f.Status), f.ReservationID, f.FulfillmentRef,
		f.Priority, UUIDArray(f.LineIDs), JSONMap(f.Metadata), f.CreatedAt, f.UpdatedAt,
		nullTime(f.CancelledAt), nullTime(f.CompletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *FulfillmentRepo) Update(ctx context.Context, f domain.Fulfillment) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE fulfillments SET
			warehouse_id=$1, status=$2, reservation_id=$3, fulfillment_ref=$4,
			priority=$5, line_ids=$6, metadata=$7, updated_at=$8,
			cancelled_at=$9, completed_at=$10
		WHERE id=$11 AND tenant_id=$12`,
		f.WarehouseID, string(f.Status), f.ReservationID, f.FulfillmentRef,
		f.Priority, UUIDArray(f.LineIDs), JSONMap(f.Metadata), f.UpdatedAt,
		nullTime(f.CancelledAt), nullTime(f.CompletedAt),
		f.ID, f.TenantID,
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

func (r *FulfillmentRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Fulfillment, error) {
	return scanFulfillment(r.DB.QueryRowContext(ctx, `
		SELECT `+fulfillmentColumns+` FROM fulfillments WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *FulfillmentRepo) ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Fulfillment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+fulfillmentColumns+` FROM fulfillments
		WHERE tenant_id=$1 AND order_id=$2
		ORDER BY created_at ASC`, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Fulfillment, 0)
	for rows.Next() {
		f, err := scanFulfillment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func scanFulfillment(row rowScanner) (domain.Fulfillment, error) {
	var f domain.Fulfillment
	var status string
	var lineIDs UUIDArray
	var meta JSONMap
	var cancelled, completed sql.NullTime
	err := row.Scan(
		&f.ID, &f.OrderID, &f.TenantID, &f.WarehouseID, &status, &f.ReservationID, &f.FulfillmentRef,
		&f.Priority, &lineIDs, &meta, &f.CreatedAt, &f.UpdatedAt, &cancelled, &completed,
	)
	if isNoRows(err) {
		return domain.Fulfillment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Fulfillment{}, err
	}
	f.Status = domain.FulfillmentStatus(status)
	f.LineIDs = []uuid.UUID(lineIDs)
	f.Metadata = map[string]any(meta)
	if f.Metadata == nil {
		f.Metadata = map[string]any{}
	}
	f.CancelledAt = scanNullTime(cancelled)
	f.CompletedAt = scanNullTime(completed)
	f.CreatedAt = f.CreatedAt.UTC()
	f.UpdatedAt = f.UpdatedAt.UTC()
	return f, nil
}
