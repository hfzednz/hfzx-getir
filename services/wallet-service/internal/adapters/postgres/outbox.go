package postgres

import (
	"context"
	"database/sql"

	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
)

// OutboxRepo persists transactional outbox rows.
type OutboxRepo struct{ DB *sql.DB }

func (o *OutboxRepo) Enqueue(ctx context.Context, m domain.OutboxMessage) error {
	payload := JSONMap(m.Payload)
	_, err := o.DB.ExecContext(ctx, `
		INSERT INTO wallet_outbox
		  (id, tenant_id, wallet_id, topic, key, payload, status, attempts, last_error,
		   created_at, updated_at, published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TenantID, m.WalletID, m.Topic, m.Key, payload, string(m.Status), m.Attempts, m.LastError,
		m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	return err
}

func (o *OutboxRepo) Update(ctx context.Context, m domain.OutboxMessage) error {
	payload := JSONMap(m.Payload)
	res, err := o.DB.ExecContext(ctx, `
		UPDATE wallet_outbox
		SET topic=$2, key=$3, payload=$4, status=$5, attempts=$6, last_error=$7,
		    updated_at=$8, published_at=$9
		WHERE id=$1`,
		m.ID, m.Topic, m.Key, payload, string(m.Status), m.Attempts, m.LastError,
		m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (o *OutboxRepo) ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := o.DB.QueryContext(ctx, `
		SELECT id, tenant_id, wallet_id, topic, key, payload, status, attempts, last_error,
		       created_at, updated_at, published_at
		FROM wallet_outbox WHERE status='pending'
		ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OutboxMessage
	for rows.Next() {
		var m domain.OutboxMessage
		var status string
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.WalletID, &m.Topic, &m.Key, &payload, &status, &m.Attempts, &m.LastError,
			&m.CreatedAt, &m.UpdatedAt, &published); err != nil {
			return nil, err
		}
		m.Status = domain.OutboxStatus(status)
		m.Payload = map[string]any(payload)
		m.CreatedAt = m.CreatedAt.UTC()
		m.UpdatedAt = m.UpdatedAt.UTC()
		m.PublishedAt = scanNullTime(published)
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)
