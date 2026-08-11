package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topic checkout.lifecycle.
const (
	EventCheckoutStarted   = "CheckoutStarted"
	EventCheckoutValidated = "CheckoutValidated"
	EventCheckoutCompleted = "CheckoutCompleted"
	EventCheckoutFailed    = "CheckoutFailed"
	EventCheckoutAbandoned = "CheckoutAbandoned"
	EventCheckoutRecovered = "CheckoutRecovered"
)

// Kafka topic names owned by this service.
const (
	TopicCheckoutLifecycle = "checkout.lifecycle"
)

// DomainEvent is the envelope for outbound integration / timeline events.
type DomainEvent struct {
	EventID     uuid.UUID
	Type        string
	OccurredAt  time.Time
	TenantID    uuid.UUID
	SessionID   uuid.UUID
	ActorID     *uuid.UUID
	ActorType   string
	TraceID     string
	Payload     map[string]any
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
	if e.SessionID == uuid.Nil {
		return fmt.Errorf("%w: session_id required", ErrInvalidArgument)
	}
	return nil
}

// NewDomainEvent constructs an envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, sessionID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:    uuid.New(),
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		TenantID:   tenantID,
		SessionID:  sessionID,
		Payload:    payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventCheckoutStarted, EventCheckoutValidated, EventCheckoutCompleted,
		EventCheckoutFailed, EventCheckoutAbandoned, EventCheckoutRecovered:
		return TopicCheckoutLifecycle
	default:
		return TopicCheckoutLifecycle
	}
}

// SessionEvent is the append-only timeline row projection.
type SessionEvent struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	ActorType  string
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e SessionEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: event id required", ErrInvalidArgument)
	}
	if e.SessionID == uuid.Nil {
		return fmt.Errorf("%w: session_id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: event type required", ErrInvalidArgument)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at required", ErrInvalidArgument)
	}
	return nil
}

// FromDomainEvent maps an outbound DomainEvent to a timeline SessionEvent.
func FromDomainEvent(e DomainEvent) SessionEvent {
	return SessionEvent{
		ID:         e.EventID,
		SessionID:  e.SessionID,
		TenantID:   e.TenantID,
		Type:       e.Type,
		Payload:    e.Payload,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		OccurredAt: e.OccurredAt,
		CreatedAt:  time.Now().UTC(),
	}
}
