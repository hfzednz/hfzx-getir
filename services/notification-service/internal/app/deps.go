package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// Deps aggregates application ports for notification use cases.
type Deps struct {
	Templates   ports.TemplateRepo
	Messages    ports.MessageRepo
	Preferences ports.PreferenceRepo
	Devices     ports.DeviceRepo
	Inbox       ports.InboxRepo
	Schedules   ports.ScheduleRepo
	Deliveries  ports.DeliveryRepo
	Outbox      ports.OutboxRepository
	Push        ports.PushProvider
	Email       ports.EmailProvider
	SMS         ports.SMSProvider
	WhatsApp    ports.WhatsAppProvider
	Publisher   ports.EventPublisher
	Clock       ports.Clock
	IDs         ports.IDGen
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

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, messageID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	_ = d.Outbox.Enqueue(ctx, domain.OutboxMessage{
		ID: d.newID(), TenantID: tenantID, MessageID: messageID,
		Topic: domain.TopicForEvent(eventType), Key: messageID.String(),
		Payload: payload, Status: domain.OutboxStatusPending,
		CreatedAt: now, UpdatedAt: now,
	})
}

func (d *Deps) emit(ctx context.Context, msg domain.Message, eventType string, extra map[string]any) {
	now := d.now()
	payload := map[string]any{
		"type": eventType, "messageId": msg.ID.String(), "tenantId": msg.TenantID.String(),
		"principalId": msg.PrincipalID.String(), "channel": string(msg.Channel),
		"priority": string(msg.Priority), "status": string(msg.Status), "occurredAt": now,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.enqueueOutbox(ctx, msg.TenantID, msg.ID, eventType, payload)
	if d.Publisher != nil {
		_ = d.Publisher.Publish(ctx, domain.TopicForEvent(eventType), msg.ID.String(), payload)
	}
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
