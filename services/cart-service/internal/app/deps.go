package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// Deps aggregates application ports for cart use cases.
type Deps struct {
	Carts     ports.CartRepository
	Events    ports.EventStore
	Outbox    ports.OutboxRepository
	Saved     ports.SavedCartRepository
	Publisher ports.EventPublisher
	Pricing   ports.PricingClient
	Inventory ports.InventoryClient
	Recommend ports.RecommendClient
	Clock     ports.Clock
	IDs       ports.IDGen
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) appendEvent(ctx context.Context, cartID, tenantID uuid.UUID, eventType string, payload map[string]any) error {
	now := d.now()
	ev := domain.CartEvent{
		ID:         d.newID(),
		CartID:     cartID,
		TenantID:   tenantID,
		Type:       eventType,
		Payload:    payload,
		OccurredAt: now,
		CreatedAt:  now,
	}
	if payload == nil {
		ev.Payload = map[string]any{}
	}
	if d.Events != nil {
		if err := d.Events.Append(ctx, ev); err != nil {
			return err
		}
	}
	d.enqueueOutbox(ctx, tenantID, cartID, eventType, map[string]any{
		"type":       eventType,
		"cartId":     cartID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    ev.Payload,
	})
	if d.Publisher != nil {
		_ = d.Publisher.Publish(ctx, domain.TopicForEvent(eventType), cartID.String(), map[string]any{
			"type":       eventType,
			"cartId":     cartID.String(),
			"tenantId":   tenantID.String(),
			"occurredAt": now,
			"payload":    ev.Payload,
		})
	}
	return nil
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, cartID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		CartID:    cartID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       cartID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
}

func (d *Deps) loadCart(ctx context.Context, tenantID, cartID uuid.UUID) (domain.Cart, error) {
	if d.Carts == nil {
		return domain.Cart{}, domain.ErrNotFound
	}
	return d.Carts.GetByID(ctx, tenantID, cartID)
}
