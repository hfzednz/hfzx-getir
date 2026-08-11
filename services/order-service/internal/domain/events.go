package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topic order.lifecycle.
const (
	EventOrderCreated      = "OrderCreated"
	EventOrderValidated    = "OrderValidated"
	EventInventoryReserved = "InventoryReserved"
	EventPaymentAuthorized = "PaymentAuthorized"
	EventWarehouseAssigned = "WarehouseAssigned"
	EventPickingStarted    = "PickingStarted"
	EventPackingCompleted  = "PackingCompleted"
	EventCourierAssigned   = "CourierAssigned"
	EventOutForDelivery    = "OutForDelivery"
	EventDelivered         = "Delivered"
	EventCompleted         = "Completed"
	EventCancelled         = "Cancelled"
	EventRefundRequested   = "RefundRequested"
	EventRefundCompleted   = "RefundCompleted"
	EventOrderArchived     = "OrderArchived"
	EventReturnRequested   = "ReturnRequested"
	EventReturnCompleted   = "ReturnCompleted"
	EventPaymentFailed     = "PaymentFailed"
	EventInventoryFailed   = "InventoryFailed"
)

// Kafka topic names owned by this service.
const (
	TopicOrderLifecycle = "order.lifecycle"
)

// DomainEvent is the envelope for outbound integration / timeline events.
// Payloads are camelCase JSON at the wire layer.
type DomainEvent struct {
	EventID    uuid.UUID
	Type       string
	OccurredAt time.Time
	TenantID   uuid.UUID
	OrderID    uuid.UUID
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
	if e.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	return nil
}

// NewDomainEvent constructs a validated-ready envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, orderID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:    uuid.New(),
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		TenantID:   tenantID,
		OrderID:    orderID,
		Payload:    payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventOrderCreated, EventOrderValidated, EventInventoryReserved,
		EventPaymentAuthorized, EventWarehouseAssigned, EventPickingStarted,
		EventPackingCompleted, EventCourierAssigned, EventOutForDelivery,
		EventDelivered, EventCompleted, EventCancelled, EventRefundRequested,
		EventRefundCompleted, EventOrderArchived, EventReturnRequested,
		EventReturnCompleted, EventPaymentFailed, EventInventoryFailed:
		return TopicOrderLifecycle
	default:
		return TopicOrderLifecycle
	}
}

// OrderEvent is the append-only timeline row projection (order_events table).
type OrderEvent struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	ActorType  string
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e OrderEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: event id required", ErrInvalidArgument)
	}
	if e.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
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

// FromDomainEvent maps an outbound DomainEvent to a timeline OrderEvent.
func FromDomainEvent(e DomainEvent) OrderEvent {
	return OrderEvent{
		ID:         e.EventID,
		OrderID:    e.OrderID,
		TenantID:   e.TenantID,
		Type:       e.Type,
		Payload:    e.Payload,
		ActorID:    e.ActorID,
		ActorType:  e.ActorType,
		OccurredAt: e.OccurredAt,
		CreatedAt:  time.Now().UTC(),
	}
}
