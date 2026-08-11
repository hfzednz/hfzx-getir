package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// EventRepo persists settlement timeline events.
type EventRepo struct{ DB *sql.DB }

func (r *EventRepo) Append(ctx context.Context, e domain.SettlementEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO settlement_events (
			id, batch_id, tenant_id, event_type, payload, actor_id, actor_type, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.BatchID, e.TenantID, e.Type, JSONMap(e.Payload), nullUUID(e.ActorID), e.ActorType,
		e.OccurredAt.UTC(), e.CreatedAt.UTC())
	return err
}

func (r *EventRepo) ListByBatch(ctx context.Context, tenantID, batchID uuid.UUID) ([]domain.SettlementEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, batch_id, tenant_id, event_type, payload, actor_id, actor_type, occurred_at, created_at
		FROM settlement_events WHERE tenant_id=$1 AND batch_id=$2 ORDER BY occurred_at ASC`, tenantID, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SettlementEvent{}
	for rows.Next() {
		var e domain.SettlementEvent
		var payload JSONMap
		var actor uuid.NullUUID
		if err := rows.Scan(
			&e.ID, &e.BatchID, &e.TenantID, &e.Type, &payload, &actor, &e.ActorType, &e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = map[string]any(payload)
		e.ActorID = scanNullUUID(actor)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// OutboxRepo persists transactional outbox rows.
type OutboxRepo struct{ DB *sql.DB }

func (r *OutboxRepo) Enqueue(ctx context.Context, m domain.OutboxMessage) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO settlement_outbox (
			id, tenant_id, batch_id, topic, message_key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.TenantID, m.BatchID, m.Topic, m.Key, JSONMap(m.Payload), string(m.Status), m.Attempts, m.LastError,
		m.CreatedAt.UTC(), m.UpdatedAt.UTC(), nullTime(m.PublishedAt))
	return err
}

func (r *OutboxRepo) Update(ctx context.Context, m domain.OutboxMessage) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE settlement_outbox SET
			topic=$2, message_key=$3, payload=$4, status=$5, attempts=$6, last_error=$7,
			updated_at=$8, published_at=$9
		WHERE id=$1`,
		m.ID, m.Topic, m.Key, JSONMap(m.Payload), string(m.Status), m.Attempts, m.LastError,
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

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, batch_id, topic, message_key, payload, status, attempts, last_error,
			created_at, updated_at, published_at
		FROM settlement_outbox WHERE status='pending' ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.OutboxMessage{}
	for rows.Next() {
		var m domain.OutboxMessage
		var status string
		var payload JSONMap
		var published sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.BatchID, &m.Topic, &m.Key, &payload, &status, &m.Attempts, &m.LastError,
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

// Repos groups settlement persistence adapters.
type Repos struct {
	Batches *BatchRepo
	Events  *EventRepo
	Outbox  *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Batches: &BatchRepo{DB: db},
		Events:  &EventRepo{DB: db},
		Outbox:  &OutboxRepo{DB: db},
	}
}

var (
	_ ports.EventStore       = (*EventRepo)(nil)
	_ ports.OutboxRepository = (*OutboxRepo)(nil)
)
