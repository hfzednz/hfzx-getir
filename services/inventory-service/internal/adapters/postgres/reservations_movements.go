package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

type ReservationRepo struct{ DB *sql.DB }

func (r *ReservationRepo) Create(ctx context.Context, res domain.Reservation) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO reservations (
			id, tenant_id, warehouse_id, type, status, expires_at, priority, external_ref, actor_id,
			metadata, created_at, updated_at, released_at, consumed_at
		) VALUES ($1,$2,$3,$4::reservation_type,$5::reservation_status,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		res.ID, res.TenantID, nullUUID(res.WarehouseID), string(res.Type), string(res.Status), nullTime(res.ExpiresAt),
		res.Priority, res.ExternalRef, nullUUID(res.ActorID), JSONMap(res.Metadata), res.CreatedAt, res.UpdatedAt,
		nullTime(res.ReleasedAt), nullTime(res.ConsumedAt)); err != nil {
		return err
	}
	if err := replaceReservationLines(ctx, tx, res.ID, res.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReservationRepo) Update(ctx context.Context, res domain.Reservation) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `
		UPDATE reservations SET warehouse_id=$2, type=$3::reservation_type, status=$4::reservation_status,
			expires_at=$5, priority=$6, external_ref=$7, actor_id=$8, metadata=$9, updated_at=$10,
			released_at=$11, consumed_at=$12
		WHERE id=$1 AND tenant_id=$13`,
		res.ID, nullUUID(res.WarehouseID), string(res.Type), string(res.Status),
		nullTime(res.ExpiresAt), res.Priority, res.ExternalRef, nullUUID(res.ActorID), JSONMap(res.Metadata), res.UpdatedAt,
		nullTime(res.ReleasedAt), nullTime(res.ConsumedAt), res.TenantID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	if err := replaceReservationLines(ctx, tx, res.ID, res.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ReservationRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Reservation, error) {
	res, err := scanReservation(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, type::text, status::text, expires_at, priority, external_ref, actor_id,
			metadata, created_at, updated_at, released_at, consumed_at
		FROM reservations WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Reservation{}, mapNotFound(err)
	}
	lines, err := listReservationLines(ctx, r.DB, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	res.Lines = lines
	return res, nil
}

func (r *ReservationRepo) ListByExternalRef(ctx context.Context, tenantID uuid.UUID, externalRef string) ([]domain.Reservation, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, type::text, status::text, expires_at, priority, external_ref, actor_id,
			metadata, created_at, updated_at, released_at, consumed_at
		FROM reservations WHERE tenant_id=$1 AND external_ref=$2 ORDER BY created_at DESC`, tenantID, externalRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Reservation{}
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		lines, err := listReservationLines(ctx, r.DB, res.ID)
		if err != nil {
			return nil, err
		}
		res.Lines = lines
		out = append(out, res)
	}
	return out, rows.Err()
}

func replaceReservationLines(ctx context.Context, tx *sql.Tx, reservationID uuid.UUID, lines []domain.ReservationLine) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM reservation_lines WHERE reservation_id=$1`, reservationID); err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO reservation_lines (
				id, reservation_id, warehouse_id, variant_id, sku_code, qty, lot_id, balance_id, location_id, metadata, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			line.ID, reservationID, line.WarehouseID, line.VariantID, line.SKUCode, line.Qty,
			nullUUID(line.LotID), nullUUID(line.BalanceID), nullUUID(line.LocationID), JSONMap(line.Metadata), line.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func listReservationLines(ctx context.Context, db *sql.DB, reservationID uuid.UUID) ([]domain.ReservationLine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, reservation_id, warehouse_id, variant_id, sku_code, qty, lot_id, balance_id, location_id, metadata, created_at
		FROM reservation_lines WHERE reservation_id=$1 ORDER BY created_at`, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ReservationLine{}
	for rows.Next() {
		var line domain.ReservationLine
		var lot, bal, loc uuid.NullUUID
		var meta JSONMap
		if err := rows.Scan(&line.ID, &line.ReservationID, &line.WarehouseID, &line.VariantID, &line.SKUCode, &line.Qty,
			&lot, &bal, &loc, &meta, &line.CreatedAt); err != nil {
			return nil, err
		}
		line.LotID = scanNullUUID(lot)
		line.BalanceID = scanNullUUID(bal)
		line.LocationID = scanNullUUID(loc)
		line.Metadata = map[string]any(meta)
		out = append(out, line)
	}
	return out, rows.Err()
}

type reservationScanner interface {
	Scan(dest ...any) error
}

func scanReservation(s reservationScanner) (domain.Reservation, error) {
	var res domain.Reservation
	var wh, actor uuid.NullUUID
	var typ, status string
	var expires, released, consumed sql.NullTime
	var meta JSONMap
	err := s.Scan(&res.ID, &res.TenantID, &wh, &typ, &status, &expires, &res.Priority, &res.ExternalRef, &actor,
		&meta, &res.CreatedAt, &res.UpdatedAt, &released, &consumed)
	if err != nil {
		return domain.Reservation{}, err
	}
	res.WarehouseID = scanNullUUID(wh)
	res.Type = domain.ReservationType(typ)
	res.Status = domain.ReservationStatus(status)
	res.ExpiresAt = scanNullTime(expires)
	res.ActorID = scanNullUUID(actor)
	res.Metadata = map[string]any(meta)
	res.ReleasedAt = scanNullTime(released)
	res.ConsumedAt = scanNullTime(consumed)
	return res, nil
}

var _ ports.ReservationRepository = (*ReservationRepo)(nil)

type MovementRepo struct{ DB *sql.DB }

func (r *MovementRepo) Create(ctx context.Context, m domain.Movement) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO stock_movements (
			id, tenant_id, warehouse_id, balance_id, variant_id, sku_code, location_id, lot_id, type, qty,
			on_hand_after, reserved_after, blocked_after, incoming_after, in_transit_after, idempotency_key,
			actor_id, reason, external_ref, reservation_id, transfer_id, metadata, occurred_at, created_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9::movement_type,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
		)`,
		m.ID, m.TenantID, m.WarehouseID, nullUUID(m.BalanceID), m.VariantID, m.SKUCode, nullUUID(m.LocationID), nullUUID(m.LotID),
		string(m.Type), m.Qty, m.OnHandAfter, m.ReservedAfter, m.BlockedAfter, m.IncomingAfter, m.InTransitAfter, m.IdempotencyKey,
		nullUUID(m.ActorID), m.Reason, m.ExternalRef, nullUUID(m.ReservationID), nullUUID(m.TransferID), JSONMap(m.Metadata),
		m.OccurredAt, m.CreatedAt)
	return mapUniqueViolation(err)
}

func (r *MovementRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.Movement, error) {
	m, err := scanMovement(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, balance_id, variant_id, sku_code, location_id, lot_id, type::text, qty,
			on_hand_after, reserved_after, blocked_after, incoming_after, in_transit_after, idempotency_key,
			actor_id, reason, external_ref, reservation_id, transfer_id, metadata, occurred_at, created_at
		FROM stock_movements WHERE idempotency_key=$1`, key))
	if err != nil {
		return domain.Movement{}, mapNotFound(err)
	}
	return m, nil
}

func (r *MovementRepo) ListByBalance(ctx context.Context, balanceID uuid.UUID, limit, offset int) ([]domain.Movement, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM stock_movements WHERE balance_id=$1`, balanceID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, balance_id, variant_id, sku_code, location_id, lot_id, type::text, qty,
			on_hand_after, reserved_after, blocked_after, incoming_after, in_transit_after, idempotency_key,
			actor_id, reason, external_ref, reservation_id, transfer_id, metadata, occurred_at, created_at
		FROM stock_movements WHERE balance_id=$1 ORDER BY occurred_at DESC LIMIT $2 OFFSET $3`, balanceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Movement{}
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

type movementScanner interface {
	Scan(dest ...any) error
}

func scanMovement(s movementScanner) (domain.Movement, error) {
	var m domain.Movement
	var bal, loc, lot, actor, reservation, transfer uuid.NullUUID
	var typ string
	var meta JSONMap
	err := s.Scan(&m.ID, &m.TenantID, &m.WarehouseID, &bal, &m.VariantID, &m.SKUCode, &loc, &lot, &typ, &m.Qty,
		&m.OnHandAfter, &m.ReservedAfter, &m.BlockedAfter, &m.IncomingAfter, &m.InTransitAfter, &m.IdempotencyKey,
		&actor, &m.Reason, &m.ExternalRef, &reservation, &transfer, &meta, &m.OccurredAt, &m.CreatedAt)
	if err != nil {
		return domain.Movement{}, err
	}
	m.BalanceID = scanNullUUID(bal)
	m.LocationID = scanNullUUID(loc)
	m.LotID = scanNullUUID(lot)
	m.Type = domain.MovementType(typ)
	m.ActorID = scanNullUUID(actor)
	m.ReservationID = scanNullUUID(reservation)
	m.TransferID = scanNullUUID(transfer)
	m.Metadata = map[string]any(meta)
	return m, nil
}

var _ ports.MovementRepository = (*MovementRepo)(nil)
