package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topic cart.lifecycle.
const (
	EventCartCreated    = "CartCreated"
	EventCartUpdated    = "CartUpdated"
	EventItemAdded      = "ItemAdded"
	EventItemRemoved    = "ItemRemoved"
	EventCouponApplied  = "CouponApplied"
	EventCouponRemoved  = "CouponRemoved"
	EventQuoteRefreshed = "QuoteRefreshed"
	EventCartMerged     = "CartMerged"
	EventCartAbandoned  = "CartAbandoned"
	EventCartRecovered  = "CartRecovered"
	EventCartSaved      = "CartSaved"
	EventSoftReserved   = "SoftReserved"
)

// Kafka topic names owned by this service.
const (
	TopicCartLifecycle = "cart.lifecycle"
)

// DomainEvent is the envelope for outbound integration / timeline events.
type DomainEvent struct {
	EventID    uuid.UUID
	Type       string
	OccurredAt time.Time
	TenantID   uuid.UUID
	CartID     uuid.UUID
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
	if e.CartID == uuid.Nil {
		return fmt.Errorf("%w: cart_id required", ErrInvalidArgument)
	}
	return nil
}

// NewDomainEvent constructs an envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, cartID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:    uuid.New(),
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		TenantID:   tenantID,
		CartID:     cartID,
		Payload:    payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	_ = eventType
	return TopicCartLifecycle
}

// CartEvent is the append-only timeline row projection (cart_events table).
type CartEvent struct {
	ID         uuid.UUID
	CartID     uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	ActorType  string
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e CartEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: event id required", ErrInvalidArgument)
	}
	if e.CartID == uuid.Nil {
		return fmt.Errorf("%w: cart_id required", ErrInvalidArgument)
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

// FromDomainEvent maps an outbound DomainEvent to a timeline CartEvent.
func FromDomainEvent(e DomainEvent) CartEvent {
	return CartEvent{
		ID:         e.EventID,
		CartID:     e.CartID,
		TenantID:   e.TenantID,
		Type:       e.Type,
		Payload:    e.Payload,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		OccurredAt: e.OccurredAt,
		CreatedAt:  time.Now().UTC(),
	}
}
