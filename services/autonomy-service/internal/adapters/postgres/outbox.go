package postgres

import (
	"context"
	"database/sql"

	"github.com/nexora/autonomy-service/internal/app/ports"
	"github.com/nexora/autonomy-service/internal/domain"
)

type OutboxRepo struct{ DB *sql.DB }

var _ ports.OutboxRepository = (*OutboxRepo)(nil)

func (r *OutboxRepo) Enqueue(ctx context.Context, m domain.OutboxMessage) error {
	status := m.Status
	if status == "" {
		status = domain.OutboxStatusPending
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO auto_outbox (
			id, tenant_id, aggregate_id, topic, key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TenantID, m.AggregateID, m.Topic, m.Key, JSONMap(m.Payload), status,
		m.Attempts, m.LastError, m.CreatedAt, m.UpdatedAt, nullTime(m.PublishedAt),
	)
	return mapUniqueViolation(err)
}

func (r *OutboxRepo) Update(ctx context.Context, m domain.OutboxMessage) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE auto_outbox SET
			topic=$1, key=$2, payload=$3, status=$4, attempts=$5, last_error=$6,
			updated_at=$7, published_at=$8
		WHERE id=$9`,
		m.Topic, m.Key, JSONMap(m.Payload), m.Status, m.Attempts, m.LastError,
		m.UpdatedAt, nullTime(m.PublishedAt), m.ID,
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

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, aggregate_id, topic, key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		FROM auto_outbox
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2`, domain.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.OutboxMessage, 0)
	for rows.Next() {
		var m domain.OutboxMessage
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.AggregateID, &m.Topic, &m.Key, &payload, &m.Status,
			&m.Attempts, &m.LastError, &m.CreatedAt, &m.UpdatedAt, &published,
		); err != nil {
			return nil, err
		}
		m.Payload = map[string]any(payload)
		if m.Payload == nil {
			m.Payload = map[string]any{}
		}
		m.PublishedAt = scanNullTime(published)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}
