package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// GetDeliveryResult aggregates message + attempts.
type GetDeliveryResult struct {
	Message  domain.Message            `json:"message"`
	Attempts []domain.DeliveryAttempt  `json:"attempts"`
}

// GetDelivery returns message and delivery attempts.
func (d *Deps) GetDelivery(ctx context.Context, tenantID, messageID uuid.UUID) (GetDeliveryResult, error) {
	if tenantID == uuid.Nil || messageID == uuid.Nil {
		return GetDeliveryResult{}, fmt.Errorf("%w: tenant_id and message_id required", domain.ErrInvalidArgument)
	}
	msg, err := d.Messages.Get(ctx, tenantID, messageID)
	if err != nil {
		return GetDeliveryResult{}, err
	}
	attempts, err := d.Deliveries.ListAttempts(ctx, tenantID, messageID)
	if err != nil {
		return GetDeliveryResult{}, err
	}
	return GetDeliveryResult{Message: msg, Attempts: attempts}, nil
}

// RetryFailedInput retries a failed message.
type RetryFailedInput struct {
	TenantID  uuid.UUID
	MessageID uuid.UUID
}

// RetryFailed re-dispatches a failed message (increments attempts).
func (d *Deps) RetryFailed(ctx context.Context, in RetryFailedInput) (domain.Message, error) {
	if in.TenantID == uuid.Nil || in.MessageID == uuid.Nil {
		return domain.Message{}, fmt.Errorf("%w: tenant_id and message_id required", domain.ErrInvalidArgument)
	}
	msg, err := d.Messages.Get(ctx, in.TenantID, in.MessageID)
	if err != nil {
		return domain.Message{}, err
	}
	if msg.Status != domain.MessageFailed && msg.Status != domain.MessageQueued {
		return domain.Message{}, fmt.Errorf("%w: message not retryable (status=%s)", domain.ErrConflict, msg.Status)
	}
	if msg.Attempts >= msg.MaxAttempts {
		_, _ = d.MoveToDLQ(ctx, MoveToDLQInput{TenantID: msg.TenantID, MessageID: msg.ID, Reason: "max_retries"})
		return msg, domain.ErrMaxRetries
	}
	return d.dispatchMessage(ctx, msg)
}

// MoveToDLQInput moves a message to DLQ.
type MoveToDLQInput struct {
	TenantID  uuid.UUID
	MessageID uuid.UUID
	Reason    string
}

// MoveToDLQ records a dead-letter entry.
func (d *Deps) MoveToDLQ(ctx context.Context, in MoveToDLQInput) (domain.DLQItem, error) {
	if in.TenantID == uuid.Nil || in.MessageID == uuid.Nil {
		return domain.DLQItem{}, fmt.Errorf("%w: tenant_id and message_id required", domain.ErrInvalidArgument)
	}
	msg, err := d.Messages.Get(ctx, in.TenantID, in.MessageID)
	if err != nil {
		return domain.DLQItem{}, err
	}
	reason := in.Reason
	if reason == "" {
		reason = msg.LastError
	}
	item := domain.DLQItem{
		ID: d.newID(), TenantID: in.TenantID, MessageID: in.MessageID,
		Reason: reason, Payload: map[string]any{
			"channel": string(msg.Channel), "priority": string(msg.Priority),
			"attempts": msg.Attempts, "lastError": msg.LastError,
		},
		CreatedAt: d.now(),
	}
	if err := d.Deliveries.MoveToDLQ(ctx, item); err != nil {
		return domain.DLQItem{}, err
	}
	msg.Status = domain.MessageFailed
	msg.UpdatedAt = d.now()
	_ = d.Messages.Update(ctx, msg)
	return item, nil
}
