package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// Deps aggregates application ports for order use cases.
type Deps struct {
	Orders       ports.OrderRepository
	Events       ports.EventStore
	Sagas        ports.SagaRepository
	Outbox       ports.OutboxRepository
	Fulfillments ports.FulfillmentRepository
	Returns      ports.ReturnRepository
	Refunds      ports.RefundRepository
	Search       ports.SearchIndexer
	Publisher    ports.EventPublisher
	Inventory    ports.InventoryClient
	Payment      ports.PaymentClient
	Warehouse    ports.WarehouseClient
	Dispatch     ports.DispatchClient
	Clock        ports.Clock
	IDs          ports.IDGen

	// PlaceLock serializes PlaceOrder by idempotency key (optional; memory provides).
	PlaceLock PlaceLocker
}

// PlaceLocker provides per-key mutual exclusion for concurrent PlaceOrder.
type PlaceLocker interface {
	WithLock(ctx context.Context, key string, fn func() error) error
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

func (d *Deps) withPlaceLock(ctx context.Context, key string, fn func() error) error {
	if d.PlaceLock != nil && key != "" {
		return d.PlaceLock.WithLock(ctx, key, fn)
	}
	return fn()
}

func (d *Deps) appendEvent(ctx context.Context, orderID, tenantID uuid.UUID, eventType string, payload map[string]any) error {
	now := d.now()
	ev := domain.OrderEvent{
		ID:         d.newID(),
		OrderID:    orderID,
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
	d.enqueueOutbox(ctx, tenantID, orderID, eventType, map[string]any{
		"type":       eventType,
		"orderId":    orderID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    ev.Payload,
	})
	return nil
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, orderID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		OrderID:   orderID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       orderID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
}

func (d *Deps) indexOrder(ctx context.Context, o domain.Order) {
	if d.Search == nil {
		return
	}
	skus := make([]string, 0, len(o.Lines))
	for _, l := range o.Lines {
		if l.SKUCode != "" {
			skus = append(skus, l.SKUCode)
		}
	}
	_ = d.Search.IndexOrder(ctx, ports.SearchDocument{
		OrderID:             o.ID,
		TenantID:            o.TenantID,
		CustomerPrincipalID: o.CustomerPrincipalID,
		Status:              string(o.Status),
		Type:                string(o.Type),
		Currency:            o.Currency,
		TotalMinor:          o.TotalMinor,
		Priority:            o.Priority,
		IdempotencyKey:      o.IdempotencyKey,
		SKUCodes:            skus,
		WarehouseIDs:        o.WarehouseIDs,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
	})
}

func (d *Deps) transition(o *domain.Order, next domain.OrderStatus) error {
	if err := domain.ValidateTransition(o.Status, next); err != nil {
		return err
	}
	if o.Status == next {
		return nil
	}
	now := d.now()
	o.Status = next
	o.UpdatedAt = now
	o.Version++
	switch next {
	case domain.OrderStatusPendingPayment:
		if o.PlacedAt == nil {
			o.PlacedAt = &now
		}
	case domain.OrderStatusCancelled:
		o.CancelledAt = &now
	case domain.OrderStatusCompleted:
		o.CompletedAt = &now
	case domain.OrderStatusArchived:
		o.ArchivedAt = &now
	}
	return nil
}
