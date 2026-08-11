package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// DispatchRepo persists dispatches, events, attempts, and batches.
type DispatchRepo struct{ DB *sql.DB }

func (r *DispatchRepo) Create(ctx context.Context, d domain.Dispatch) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO dispatches (
			id, tenant_id, order_id, fulfillment_id, warehouse_id, courier_principal_id, vehicle_id,
			status, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng, required_vehicle, batch_id, route_id,
			eta_seconds, pod_type, pod_reference, fail_reason, fail_note, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22
		)`,
		d.ID, d.TenantID, d.OrderID, nullUUIDValue(d.FulfillmentID), nullUUIDValue(d.WarehouseID),
		nullUUID(d.CourierPrincipalID), nullUUID(d.VehicleID), string(d.Status),
		d.Pickup.Lat, d.Pickup.Lng, d.Dropoff.Lat, d.Dropoff.Lng, string(d.RequiredVehicle),
		nullUUID(d.BatchID), nullUUID(d.RouteID), nullInt(d.ETASeconds),
		nullPOD(d.PODType), d.PODReference, nullFail(d.FailReason), d.FailNote,
		d.CreatedAt.UTC(), d.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *DispatchRepo) Update(ctx context.Context, d domain.Dispatch) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE dispatches SET
			order_id=$3, fulfillment_id=$4, warehouse_id=$5, courier_principal_id=$6, vehicle_id=$7,
			status=$8, pickup_lat=$9, pickup_lng=$10, dropoff_lat=$11, dropoff_lng=$12, required_vehicle=$13,
			batch_id=$14, route_id=$15, eta_seconds=$16, pod_type=$17, pod_reference=$18, fail_reason=$19,
			fail_note=$20, updated_at=$21
		WHERE id=$1 AND tenant_id=$2`,
		d.ID, d.TenantID, d.OrderID, nullUUIDValue(d.FulfillmentID), nullUUIDValue(d.WarehouseID),
		nullUUID(d.CourierPrincipalID), nullUUID(d.VehicleID), string(d.Status),
		d.Pickup.Lat, d.Pickup.Lng, d.Dropoff.Lat, d.Dropoff.Lng, string(d.RequiredVehicle),
		nullUUID(d.BatchID), nullUUID(d.RouteID), nullInt(d.ETASeconds),
		nullPOD(d.PODType), d.PODReference, nullFail(d.FailReason), d.FailNote, d.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: dispatch %s", domain.ErrNotFound, d.ID)
	}
	return nil
}

func (r *DispatchRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	d, err := scanDispatch(r.DB.QueryRowContext(ctx, dispatchSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.Dispatch{}, fmt.Errorf("%w: dispatch %s", domain.ErrNotFound, id)
	}
	return d, err
}

func (r *DispatchRepo) List(ctx context.Context, tenantID uuid.UUID, status domain.DispatchStatus, limit, offset int) ([]domain.Dispatch, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	args := []any{tenantID}
	where := `WHERE tenant_id=$1`
	if status != "" {
		where += ` AND status=$2`
		args = append(args, string(status))
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM dispatches `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := dispatchSelect + ` ` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Dispatch{}
	for rows.Next() {
		d, err := scanDispatch(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *DispatchRepo) AppendEvent(ctx context.Context, e domain.DispatchEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO dispatch_events (id, tenant_id, dispatch_id, type, from_status, to_status, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.DispatchID, e.Type, nullStringEnum(string(e.FromStatus)),
		nullStringEnum(string(e.ToStatus)), JSONMap(e.Payload), e.CreatedAt.UTC())
	return err
}

func (r *DispatchRepo) ListEvents(ctx context.Context, tenantID, dispatchID uuid.UUID) ([]domain.DispatchEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, dispatch_id, type, from_status, to_status, payload, created_at
		FROM dispatch_events WHERE tenant_id=$1 AND dispatch_id=$2 ORDER BY created_at ASC`, tenantID, dispatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DispatchEvent{}
	for rows.Next() {
		var e domain.DispatchEvent
		var from, to sql.NullString
		var payload JSONMap
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.DispatchID, &e.Type, &from, &to, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		if from.Valid {
			e.FromStatus = domain.DispatchStatus(from.String)
		}
		if to.Valid {
			e.ToStatus = domain.DispatchStatus(to.String)
		}
		e.Payload = map[string]any(payload)
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *DispatchRepo) AppendAttempt(ctx context.Context, a domain.AssignmentAttempt) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO assignment_attempts (
			id, tenant_id, dispatch_id, courier_principal_id, strategy, success, distance_m, reason, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.TenantID, a.DispatchID, nullUUID(a.CourierPrincipalID), a.Strategy, a.Success,
		nullFloat(a.DistanceM), a.Reason, a.CreatedAt.UTC())
	return err
}

func (r *DispatchRepo) CreateBatch(ctx context.Context, b domain.Batch) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO batches (id, tenant_id, label, dispatch_ids, created_at)
		VALUES ($1,$2,$3,$4,$5)`,
		b.ID, b.TenantID, b.Label, UUIDArray(b.DispatchIDs), b.CreatedAt.UTC())
	return err
}

func (r *DispatchRepo) GetBatch(ctx context.Context, tenantID, id uuid.UUID) (domain.Batch, error) {
	var b domain.Batch
	var ids UUIDArray
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, label, dispatch_ids, created_at
		FROM batches WHERE id=$1 AND tenant_id=$2`, id, tenantID).Scan(
		&b.ID, &b.TenantID, &b.Label, &ids, &b.CreatedAt)
	if isNoRows(err) {
		return domain.Batch{}, fmt.Errorf("%w: batch %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.Batch{}, err
	}
	b.DispatchIDs = []uuid.UUID(ids)
	b.CreatedAt = b.CreatedAt.UTC()
	return b, nil
}

const dispatchSelect = `
	SELECT id, tenant_id, order_id, fulfillment_id, warehouse_id, courier_principal_id, vehicle_id,
		status, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng, required_vehicle, batch_id, route_id,
		eta_seconds, pod_type, pod_reference, fail_reason, fail_note, created_at, updated_at
	FROM dispatches`

type scanner interface {
	Scan(dest ...any) error
}

func scanDispatch(row scanner) (domain.Dispatch, error) {
	var d domain.Dispatch
	var status, vehicle string
	var fulfillment, warehouse, courier, vehicleID, batchID, routeID uuid.NullUUID
	var eta sql.NullInt64
	var pod, fail sql.NullString
	err := row.Scan(
		&d.ID, &d.TenantID, &d.OrderID, &fulfillment, &warehouse, &courier, &vehicleID,
		&status, &d.Pickup.Lat, &d.Pickup.Lng, &d.Dropoff.Lat, &d.Dropoff.Lng, &vehicle,
		&batchID, &routeID, &eta, &pod, &d.PODReference, &fail, &d.FailNote, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.Dispatch{}, err
	}
	d.Status = domain.DispatchStatus(status)
	d.RequiredVehicle = domain.VehicleType(vehicle)
	d.FulfillmentID = scanUUIDOrNil(fulfillment)
	d.WarehouseID = scanUUIDOrNil(warehouse)
	d.CourierPrincipalID = scanNullUUID(courier)
	d.VehicleID = scanNullUUID(vehicleID)
	d.BatchID = scanNullUUID(batchID)
	d.RouteID = scanNullUUID(routeID)
	d.ETASeconds = scanNullInt(eta)
	if pod.Valid && pod.String != "" {
		t := domain.PODType(pod.String)
		d.PODType = &t
	}
	if fail.Valid && fail.String != "" {
		f := domain.FailReason(fail.String)
		d.FailReason = &f
	}
	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()
	return d, nil
}

func nullUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func scanUUIDOrNil(n uuid.NullUUID) uuid.UUID {
	if !n.Valid {
		return uuid.Nil
	}
	return n.UUID
}

func nullPOD(t *domain.PODType) any {
	if t == nil || *t == "" {
		return nil
	}
	return string(*t)
}

func nullFail(r *domain.FailReason) any {
	if r == nil || *r == "" {
		return nil
	}
	return string(*r)
}

var _ ports.DispatchRepo = (*DispatchRepo)(nil)
