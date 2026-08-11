package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// DeliveryRepo persists delivery attempts, events, DLQ, and provider routes.
type DeliveryRepo struct{ DB *sql.DB }

func (r *DeliveryRepo) CreateAttempt(ctx context.Context, a domain.DeliveryAttempt) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO deliveries (
			id, tenant_id, message_id, attempt_no, provider, status, provider_ref, error, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.TenantID, a.MessageID, a.AttemptNo, a.Provider, a.Status, a.ProviderRef, a.Error, a.CreatedAt.UTC())
	return err
}

func (r *DeliveryRepo) ListAttempts(ctx context.Context, tenantID, messageID uuid.UUID) ([]domain.DeliveryAttempt, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, message_id, attempt_no, provider, status, provider_ref, error, created_at
		FROM deliveries WHERE tenant_id=$1 AND message_id=$2 ORDER BY attempt_no ASC`, tenantID, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DeliveryAttempt{}
	for rows.Next() {
		var a domain.DeliveryAttempt
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.MessageID, &a.AttemptNo, &a.Provider, &a.Status,
			&a.ProviderRef, &a.Error, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *DeliveryRepo) CreateEvent(ctx context.Context, e domain.DeliveryEvent) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO delivery_events (id, tenant_id, message_id, type, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.TenantID, e.MessageID, e.Type, JSONMap(e.Payload), e.CreatedAt.UTC())
	return err
}

func (r *DeliveryRepo) MoveToDLQ(ctx context.Context, item domain.DLQItem) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO dlq (id, tenant_id, message_id, reason, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		item.ID, item.TenantID, item.MessageID, item.Reason, JSONMap(item.Payload), item.CreatedAt.UTC())
	return err
}

func (r *DeliveryRepo) ListDLQ(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.DLQItem, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, message_id, reason, payload, created_at
		FROM dlq WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DLQItem{}
	for rows.Next() {
		var item domain.DLQItem
		var payload JSONMap
		if err := rows.Scan(&item.ID, &item.TenantID, &item.MessageID, &item.Reason, &payload, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Payload = map[string]any(payload)
		item.CreatedAt = item.CreatedAt.UTC()
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *DeliveryRepo) UpsertRoute(ctx context.Context, route domain.ProviderRoute) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO provider_routes (id, tenant_id, channel, provider, priority, enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, channel, provider) DO UPDATE SET
			id=EXCLUDED.id, priority=EXCLUDED.priority, enabled=EXCLUDED.enabled, updated_at=EXCLUDED.updated_at`,
		route.ID, route.TenantID, string(route.Channel), route.Provider, route.Priority, route.Enabled,
		route.CreatedAt.UTC(), route.UpdatedAt.UTC())
	return err
}

func (r *DeliveryRepo) ListRoutes(ctx context.Context, tenantID uuid.UUID) ([]domain.ProviderRoute, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, channel, provider, priority, enabled, created_at, updated_at
		FROM provider_routes WHERE tenant_id=$1 ORDER BY priority DESC, provider ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProviderRoute{}
	for rows.Next() {
		var route domain.ProviderRoute
		var channel string
		if err := rows.Scan(
			&route.ID, &route.TenantID, &channel, &route.Provider, &route.Priority, &route.Enabled,
			&route.CreatedAt, &route.UpdatedAt); err != nil {
			return nil, err
		}
		route.Channel = domain.Channel(channel)
		route.CreatedAt = route.CreatedAt.UTC()
		route.UpdatedAt = route.UpdatedAt.UTC()
		out = append(out, route)
	}
	return out, rows.Err()
}

var _ ports.DeliveryRepo = (*DeliveryRepo)(nil)
