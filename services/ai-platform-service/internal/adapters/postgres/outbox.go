package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexora/ai-platform-service/internal/app/ports"
	"github.com/nexora/ai-platform-service/internal/domain"
)

// OutboxRepo persists transactional outbox rows.
type OutboxRepo struct{ DB *sql.DB }

func (r *OutboxRepo) Enqueue(ctx context.Context, msg domain.OutboxMessage) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO outbox_messages (
			id, tenant_id, aggregate_id, topic, key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		msg.ID, msg.TenantID, msg.AggregateID, msg.Topic, msg.Key, JSONMap(msg.Payload),
		msg.Status, msg.Attempts, msg.LastError, msg.CreatedAt.UTC(), msg.UpdatedAt.UTC(),
		nullTime(msg.PublishedAt))
	return err
}

func (r *OutboxRepo) Update(ctx context.Context, msg domain.OutboxMessage) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE outbox_messages SET
			topic=$2, key=$3, payload=$4, status=$5, attempts=$6, last_error=$7,
			updated_at=$8, published_at=$9
		WHERE id=$1`,
		msg.ID, msg.Topic, msg.Key, JSONMap(msg.Payload), msg.Status, msg.Attempts, msg.LastError,
		msg.UpdatedAt.UTC(), nullTime(msg.PublishedAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("%w: outbox", domain.ErrNotFound)
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
		FROM outbox_messages WHERE status=$1 ORDER BY created_at ASC LIMIT $2`,
		domain.OutboxStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OutboxMessage{}
	for rows.Next() {
		var m domain.OutboxMessage
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.AggregateID, &m.Topic, &m.Key, &payload, &m.Status, &m.Attempts, &m.LastError,
			&m.CreatedAt, &m.UpdatedAt, &published); err != nil {
			return nil, err
		}
		m.Payload = map[string]any(payload)
		m.PublishedAt = scanNullTime(published)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)
