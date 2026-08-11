package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// Deps aggregates application ports for checkout use cases.
type Deps struct {
	Sessions   ports.CheckoutRepo
	Events     ports.EventStore
	Outbox     ports.OutboxRepository
	Publisher  ports.EventPublisher
	Cart       ports.CartClient
	Pricing    ports.PricingClient
	Inventory  ports.InventoryClient
	Geofence   ports.GeofenceClient
	Fraud      ports.FraudClient
	PayElig    ports.PaymentEligibilityClient
	Orders     ports.OrderClient
	Promo      ports.PromoClient
	Customer   ports.CustomerClient
	Clock      ports.Clock
	IDs        ports.IDGen

	// CompleteLock serializes Complete by idempotency key (optional).
	CompleteLock CompleteLocker

	// MinOrderMinor is a fallback minimum order when geofence does not supply one.
	MinOrderMinor int64

	// AutoPlaceOrder when true calls OrderClient.PlaceOrder after create (optional).
	AutoPlaceOrder bool
}

// CompleteLocker provides per-key mutual exclusion for concurrent Complete.
type CompleteLocker interface {
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

func (d *Deps) withCompleteLock(ctx context.Context, key string, fn func() error) error {
	if d.CompleteLock != nil && key != "" {
		return d.CompleteLock.WithLock(ctx, key, fn)
	}
	return fn()
}

func (d *Deps) appendEvent(ctx context.Context, sessionID, tenantID uuid.UUID, eventType string, payload map[string]any) error {
	now := d.now()
	ev := domain.SessionEvent{
		ID:         d.newID(),
		SessionID:  sessionID,
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
	d.enqueueOutbox(ctx, tenantID, sessionID, eventType, map[string]any{
		"type":       eventType,
		"sessionId":  sessionID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    ev.Payload,
	})
	return nil
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, sessionID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		SessionID: sessionID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       sessionID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
}

func (d *Deps) transition(s *domain.Session, next domain.SessionStatus) error {
	if err := domain.ValidateTransition(s.Status, next); err != nil {
		return err
	}
	if s.Status == next {
		return nil
	}
	now := d.now()
	s.Status = next
	s.UpdatedAt = now
	s.Version++
	switch next {
	case domain.StatusCompleted:
		s.CompletedAt = &now
	case domain.StatusAbandoned:
		s.AbandonedAt = &now
	case domain.StatusFailed:
		s.FailedAt = &now
	}
	return nil
}

// PublishPending drains the outbox via EventPublisher.
func (d *Deps) PublishPending(ctx context.Context, limit int) (int, error) {
	if d.Outbox == nil || d.Publisher == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	msgs, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	now := d.now()
	for _, m := range msgs {
		if err := d.Publisher.Publish(ctx, m.Topic, m.Key, m.Payload); err != nil {
			m.Attempts++
			m.LastError = err.Error()
			m.Status = domain.OutboxStatusFailed
			m.UpdatedAt = now
			_ = d.Outbox.Update(ctx, m)
			continue
		}
		m.Attempts++
		m.Status = domain.OutboxStatusPublished
		m.PublishedAt = &now
		m.UpdatedAt = now
		_ = d.Outbox.Update(ctx, m)
		n++
	}
	return n, nil
}
