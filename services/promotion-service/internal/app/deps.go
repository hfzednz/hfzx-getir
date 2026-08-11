package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app/ports"
	"github.com/nexora/promotion-service/internal/domain"
)

// Deps aggregates application ports for promotion use cases.
type Deps struct {
	Campaigns   ports.CampaignRepository
	Promotions  ports.PromotionRepository
	Rules       ports.RuleRepository
	Coupons     ports.CouponRepository
	Vouchers    ports.VoucherRepository
	Usage       ports.UsageRepository
	Simulations ports.SimulationRepository
	Outbox      ports.OutboxRepository
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

func (d *Deps) emit(ctx context.Context, tenantID, aggregateID uuid.UUID, eventType string, payload map[string]any) {
	now := d.now()
	if payload == nil {
		payload = map[string]any{}
	}
	body := map[string]any{
		"type":       eventType,
		"aggregateId": aggregateID.String(),
		"tenantId":   tenantID.String(),
		"occurredAt": now,
		"payload":    payload,
	}
	if d.Outbox != nil {
		_ = d.Outbox.Enqueue(ctx, domain.OutboxMessage{
			ID:          d.newID(),
			TenantID:    tenantID,
			AggregateID: aggregateID,
			Topic:       domain.TopicForEvent(eventType),
			Key:         aggregateID.String(),
			Payload:     body,
			Status:      domain.OutboxStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	if d.Publisher != nil {
		_ = d.Publisher.Publish(ctx, domain.TopicForEvent(eventType), aggregateID.String(), body)
	}
}
