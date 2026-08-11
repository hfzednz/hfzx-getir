package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topic payment.lifecycle.
const (
	EventPaymentInitiated   = "PaymentInitiated"
	EventPaymentAuthorized  = "PaymentAuthorized"
	EventPaymentCaptured    = "PaymentCaptured"
	EventPaymentFailed      = "PaymentFailed"
	EventPaymentVoided      = "PaymentVoided"
	EventRefundRequested    = "RefundRequested"
	EventRefundCompleted    = "RefundCompleted"
	EventChargebackCreated  = "ChargebackCreated"
)

// Kafka topic names owned by this service.
const (
	TopicPaymentLifecycle = "payment.lifecycle"
)

// DomainEvent is the envelope for outbound integration events.
type DomainEvent struct {
	EventID    uuid.UUID
	Type       string
	OccurredAt time.Time
	TenantID   uuid.UUID
	IntentID   uuid.UUID
	ActorID    *uuid.UUID
	ActorType  string
	TraceID    string
	Payload    map[string]any
}

// Validate checks structural invariants of the event envelope.
func (e DomainEvent) Validate() error {
	if e.EventID == uuid.Nil {
		return fmt.Errorf("%w: event_id required", ErrInvalidArgument)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: event type required", ErrInvalidArgument)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.IntentID == uuid.Nil {
		return fmt.Errorf("%w: intent_id required", ErrInvalidArgument)
	}
	return nil
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicPaymentLifecycle
}
