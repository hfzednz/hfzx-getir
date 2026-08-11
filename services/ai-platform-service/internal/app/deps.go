package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/app/ports"
	"github.com/nexora/ai-platform-service/internal/domain"
)

type Deps struct {
	Features   ports.FeatureRepo
	Models     ports.ModelRepo
	Prompts    ports.PromptRepo
	Memory     ports.MemoryRepo
	Agents     ports.AgentRepo
	Automation ports.AutomationRepo
	Drift      ports.DriftRepo
	Outbox     ports.OutboxRepository
	Runtime    ports.InferenceRuntime
	LLM        ports.LLMProvider
	RAG        ports.RAGRetriever
	Embed      ports.VectorEmbedder
	Publisher  ports.EventPublisher
	Clock      ports.Clock
	IDs        ports.IDGen
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

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

func (d *Deps) emit(ctx context.Context, tenantID, aggregateID uuid.UUID, eventType string, extra map[string]any) {
	now := d.now()
	payload := map[string]any{
		"type": eventType, "aggregateId": aggregateID.String(),
		"tenantId": tenantID.String(), "occurredAt": now,
	}
	for k, v := range extra {
		payload[k] = v
	}
	if d.Outbox != nil {
		_ = d.Outbox.Enqueue(ctx, domain.OutboxMessage{
			ID: d.newID(), TenantID: tenantID, AggregateID: aggregateID,
			Topic: domain.TopicForEvent(eventType), Key: aggregateID.String(),
			Payload: payload, Status: domain.OutboxStatusPending,
			CreatedAt: now, UpdatedAt: now,
		})
	}
	if d.Publisher != nil {
		_ = d.Publisher.Publish(ctx, domain.TopicForEvent(eventType), aggregateID.String(), payload)
	}
}

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
