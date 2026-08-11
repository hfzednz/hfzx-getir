package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// EventStoreRepo persists append-only checkout timeline events.
type EventStoreRepo struct{ DB *sql.DB }

var _ ports.EventStore = (*EventStoreRepo)(nil)

func (r *EventStoreRepo) Append(ctx context.Context, e domain.SessionEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO checkout_events (
			id, session_id, tenant_id, type, payload, actor_id, actor_type, occurred_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.SessionID, e.TenantID, e.Type, JSONMap(e.Payload),
		nullUUID(e.ActorID), e.ActorType, e.OccurredAt, e.CreatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *EventStoreRepo) ListBySession(ctx context.Context, tenantID, sessionID uuid.UUID) ([]domain.SessionEvent, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, session_id, tenant_id, type, payload, actor_id, actor_type, occurred_at, created_at
		FROM checkout_events
		WHERE tenant_id=$1 AND session_id=$2
		ORDER BY occurred_at ASC, created_at ASC`, tenantID, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SessionEvent, 0)
	for rows.Next() {
		var e domain.SessionEvent
		var payload JSONMap
		var actor uuid.NullUUID
		if err := rows.Scan(
			&e.ID, &e.SessionID, &e.TenantID, &e.Type, &payload, &actor, &e.ActorType, &e.OccurredAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.Payload = map[string]any(payload)
		if e.Payload == nil {
			e.Payload = map[string]any{}
		}
		e.ActorID = scanNullUUID(actor)
		e.OccurredAt = e.OccurredAt.UTC()
		e.CreatedAt = e.CreatedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}
