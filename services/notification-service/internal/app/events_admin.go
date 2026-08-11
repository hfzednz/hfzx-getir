package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// HandleDomainEventInput maps an inbound domain event to a template send.
type HandleDomainEventInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	OrderID     *uuid.UUID
	EventType   string
	Channel     domain.Channel
	Priority    domain.Priority
	Recipient   string
	Locale      string
	Vars        map[string]string
	IdempotencyKey string
}

// HandleDomainEvent maps OrderDelivered etc. to template keys and sends.
func (d *Deps) HandleDomainEvent(ctx context.Context, in HandleDomainEventInput) (domain.Message, error) {
	key, ok := domain.TemplateKeyForDomainEvent(in.EventType)
	if !ok {
		return domain.Message{}, fmt.Errorf("%w: unknown event type %s", domain.ErrInvalidArgument, in.EventType)
	}
	channel := in.Channel
	if channel == "" {
		channel = domain.ChannelPush
	}
	priority := in.Priority
	if priority == "" {
		priority = domain.PriorityTransactional
	}
	idem := in.IdempotencyKey
	if idem == "" {
		idem = in.EventType + ":" + in.PrincipalID.String()
		if in.OrderID != nil {
			idem = in.EventType + ":" + in.OrderID.String()
		}
	}
	vars := in.Vars
	if vars == nil {
		vars = map[string]string{}
	}
	if _, ok := vars["eventType"]; !ok {
		vars["eventType"] = in.EventType
	}
	return d.Send(ctx, SendInput{
		TenantID: in.TenantID, PrincipalID: in.PrincipalID, OrderID: in.OrderID,
		Channel: channel, Priority: priority, TemplateKey: key, Locale: in.Locale,
		Recipient: in.Recipient, Vars: vars, IdempotencyKey: idem,
	})
}

// BestSendTimeResult is an AI stub.
type BestSendTimeResult struct {
	HourUTC int    `json:"hourUtc"`
	Reason  string `json:"reason"`
}

// BestSendTime returns a stub best-send-time hint.
func (d *Deps) BestSendTime(_ context.Context, _ uuid.UUID, _ uuid.UUID) (BestSendTimeResult, error) {
	return BestSendTimeResult{HourUTC: 10, Reason: "stub: mid-morning engagement peak"}, nil
}

// RecommendChannelResult is an AI stub.
type RecommendChannelResult struct {
	Channel domain.Channel `json:"channel"`
	Reason  string         `json:"reason"`
}

// RecommendChannel returns a stub channel recommendation.
func (d *Deps) RecommendChannel(_ context.Context, _ uuid.UUID, _ uuid.UUID, priority domain.Priority) (RecommendChannelResult, error) {
	ch := domain.ChannelPush
	if priority == domain.PriorityOTP {
		ch = domain.ChannelSMS
	}
	return RecommendChannelResult{Channel: ch, Reason: "stub: default channel by priority"}, nil
}

// AdminStatsResult aggregates delivery stats.
type AdminStatsResult struct {
	ByStatus map[string]int `json:"byStatus"`
	DLQCount int            `json:"dlqCount"`
}

// AdminStats returns admin dashboard aggregates.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (AdminStatsResult, error) {
	if tenantID == uuid.Nil {
		return AdminStatsResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	counts, err := d.Messages.CountByStatus(ctx, tenantID)
	if err != nil {
		return AdminStatsResult{}, err
	}
	by := map[string]int{}
	for k, v := range counts {
		by[string(k)] = v
	}
	dlq, err := d.Deliveries.ListDLQ(ctx, tenantID, 1000)
	if err != nil {
		return AdminStatsResult{}, err
	}
	return AdminStatsResult{ByStatus: by, DLQCount: len(dlq)}, nil
}
