package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// SagaRepo persists saga instances and steps.
type SagaRepo struct{ DB *sql.DB }

var _ ports.SagaRepository = (*SagaRepo)(nil)

const sagaColumns = `
	id, order_id, tenant_id, saga_type, status, current_step, correlation_id,
	idempotency_key, last_error, metadata, started_at, completed_at, created_at, updated_at
`

func (r *SagaRepo) Create(ctx context.Context, s domain.SagaInstance) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO saga_instances (
			id, order_id, tenant_id, saga_type, status, current_step, correlation_id,
			idempotency_key, last_error, metadata, started_at, completed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		s.ID, s.OrderID, s.TenantID, string(s.SagaType), string(s.Status), s.CurrentStep, s.CorrelationID,
		s.IdempotencyKey, s.LastError, JSONMap(s.Metadata),
		nullTime(s.StartedAt), nullTime(s.CompletedAt), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return mapUniqueViolation(err)
	}
	if err := upsertSagaSteps(ctx, tx, s.Steps); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SagaRepo) Update(ctx context.Context, s domain.SagaInstance) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE saga_instances SET
			saga_type=$1, status=$2, current_step=$3, correlation_id=$4,
			last_error=$5, metadata=$6, started_at=$7, completed_at=$8, updated_at=$9
		WHERE id=$10 AND tenant_id=$11`,
		string(s.SagaType), string(s.Status), s.CurrentStep, s.CorrelationID,
		s.LastError, JSONMap(s.Metadata), nullTime(s.StartedAt), nullTime(s.CompletedAt), s.UpdatedAt,
		s.ID, s.TenantID,
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
	if _, err := tx.ExecContext(ctx, `DELETE FROM saga_steps WHERE saga_id=$1`, s.ID); err != nil {
		return err
	}
	if err := upsertSagaSteps(ctx, tx, s.Steps); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SagaRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SagaInstance, error) {
	s, err := scanSaga(r.DB.QueryRowContext(ctx, `
		SELECT `+sagaColumns+` FROM saga_instances WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.SagaInstance{}, err
	}
	steps, err := loadSagaSteps(ctx, r.DB, s.ID)
	if err != nil {
		return domain.SagaInstance{}, err
	}
	s.Steps = steps
	return s, nil
}

func (r *SagaRepo) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SagaInstance, error) {
	s, err := scanSaga(r.DB.QueryRowContext(ctx, `
		SELECT `+sagaColumns+` FROM saga_instances WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key))
	if err != nil {
		return domain.SagaInstance{}, err
	}
	steps, err := loadSagaSteps(ctx, r.DB, s.ID)
	if err != nil {
		return domain.SagaInstance{}, err
	}
	s.Steps = steps
	return s, nil
}

func (r *SagaRepo) ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.SagaInstance, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT `+sagaColumns+` FROM saga_instances
		WHERE tenant_id=$1 AND order_id=$2
		ORDER BY created_at DESC`, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SagaInstance, 0)
	for rows.Next() {
		s, err := scanSaga(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		steps, err := loadSagaSteps(ctx, r.DB, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Steps = steps
	}
	return out, nil
}

func scanSaga(row rowScanner) (domain.SagaInstance, error) {
	var s domain.SagaInstance
	var sagaType, status string
	var meta JSONMap
	var started, completed sql.NullTime
	err := row.Scan(
		&s.ID, &s.OrderID, &s.TenantID, &sagaType, &status, &s.CurrentStep, &s.CorrelationID,
		&s.IdempotencyKey, &s.LastError, &meta, &started, &completed, &s.CreatedAt, &s.UpdatedAt,
	)
	if isNoRows(err) {
		return domain.SagaInstance{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SagaInstance{}, err
	}
	s.SagaType = domain.SagaType(sagaType)
	s.Status = domain.SagaInstanceStatus(status)
	s.Metadata = map[string]any(meta)
	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}
	s.StartedAt = scanNullTime(started)
	s.CompletedAt = scanNullTime(completed)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

func upsertSagaSteps(ctx context.Context, tx *sql.Tx, steps []domain.SagaStep) error {
	for _, st := range steps {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO saga_steps (
				id, saga_id, order_id, tenant_id, name, status, attempt, last_error,
				idempotency_key, compensation_of, payload, started_at, completed_at, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
			st.ID, st.SagaID, st.OrderID, st.TenantID, st.Name, string(st.Status), st.Attempt, st.LastError,
			st.IdempotencyKey, st.CompensationOf, JSONMap(st.Payload),
			nullTime(st.StartedAt), nullTime(st.CompletedAt), st.CreatedAt, st.UpdatedAt,
		)
		if err != nil {
			return mapUniqueViolation(err)
		}
	}
	return nil
}

func loadSagaSteps(ctx context.Context, db *sql.DB, sagaID uuid.UUID) ([]domain.SagaStep, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, saga_id, order_id, tenant_id, name, status, attempt, last_error,
			idempotency_key, compensation_of, payload, started_at, completed_at, created_at, updated_at
		FROM saga_steps WHERE saga_id=$1 ORDER BY created_at ASC`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SagaStep, 0)
	for rows.Next() {
		var st domain.SagaStep
		var status string
		var payload JSONMap
		var started, completed sql.NullTime
		if err := rows.Scan(
			&st.ID, &st.SagaID, &st.OrderID, &st.TenantID, &st.Name, &status, &st.Attempt, &st.LastError,
			&st.IdempotencyKey, &st.CompensationOf, &payload, &started, &completed, &st.CreatedAt, &st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		st.Status = domain.SagaStepStatus(status)
		st.Payload = map[string]any(payload)
		if st.Payload == nil {
			st.Payload = map[string]any{}
		}
		st.StartedAt = scanNullTime(started)
		st.CompletedAt = scanNullTime(completed)
		st.CreatedAt = st.CreatedAt.UTC()
		st.UpdatedAt = st.UpdatedAt.UTC()
		out = append(out, st)
	}
	return out, rows.Err()
}
