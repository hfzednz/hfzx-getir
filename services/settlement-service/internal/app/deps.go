package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// Deps aggregates application ports for settlement use cases.
type Deps struct {
	Batches   ports.BatchRepository
	Events    ports.EventStore
	Outbox    ports.OutboxRepository
	Publisher ports.EventPublisher
	Ledger    ports.LedgerClient
	Payout    ports.PayoutClient
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

func (d *Deps) appendEvent(ctx context.Context, batchID, tenantID uuid.UUID, eventType string, payload map[string]any, actor *uuid.UUID) error {
	now := d.now()
	ev := domain.SettlementEvent{
		ID:         d.newID(),
		BatchID:    batchID,
		TenantID:   tenantID,
		Type:       eventType,
		Payload:    payload,
		ActorID:    actor,
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
	d.enqueueOutbox(ctx, tenantID, batchID, eventType, map[string]any{
		"type":       eventType,
		"batchId":    batchID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    ev.Payload,
	})
	return nil
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, batchID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	msg := domain.OutboxMessage{
		ID:        d.newID(),
		TenantID:  tenantID,
		BatchID:   batchID,
		Topic:     domain.TopicForEvent(eventType),
		Key:       batchID.String(),
		Payload:   payload,
		Status:    domain.OutboxStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = d.Outbox.Enqueue(ctx, msg)
}

func (d *Deps) transition(b *domain.SettlementBatch, next domain.BatchStatus) error {
	if err := domain.ValidateTransition(b.Status, next); err != nil {
		return err
	}
	if b.Status == next {
		return nil
	}
	now := d.now()
	b.Status = next
	b.UpdatedAt = now
	b.Version++
	switch next {
	case domain.BatchStatusCompleted:
		b.CompletedAt = &now
	case domain.BatchStatusFailed:
		b.FailedAt = &now
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
