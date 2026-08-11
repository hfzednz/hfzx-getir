package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// Deps aggregates application ports for ledger use cases.
type Deps struct {
	Accounts  ports.AccountRepository
	Journals  ports.JournalRepository
	Invoices  ports.InvoiceRepository
	TaxRules  ports.TaxRuleRepository
	Events    ports.EventStore
	Outbox    ports.OutboxRepository
	Publisher ports.EventPublisher
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

func (d *Deps) appendEvent(ctx context.Context, entityID, tenantID uuid.UUID, eventType string, payload map[string]any) error {
	now := d.now()
	ev := domain.LedgerEvent{
		ID:         d.newID(),
		EntityID:   entityID,
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
	d.enqueueOutbox(ctx, tenantID, entityID, eventType, map[string]any{
		"type":       eventType,
		"entityId":   entityID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    ev.Payload,
	})
	return nil
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, entityID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		EntityID:  entityID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       entityID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
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
