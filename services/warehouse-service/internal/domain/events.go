package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Kafka topics owned by warehouse-service.
const (
	TopicFulfillmentEvents = "warehouse.fulfillment.events"
	TopicTaskEvents        = "warehouse.task.events"
)

// Domain event type constants (ARCHITECTURE Kafka catalog).
const (
	EventOrderReceived     = "OrderReceived"
	EventPickingStarted    = "PickingStarted"
	EventPickingCompleted  = "PickingCompleted"
	EventPackingStarted    = "PackingStarted"
	EventPackingCompleted  = "PackingCompleted"
	EventLabelGenerated    = "LabelGenerated"
	EventDispatchStarted   = "DispatchStarted"
	EventDispatchCompleted = "DispatchCompleted"
	EventCourierAssigned   = "CourierAssigned"
	EventTaskAssigned      = "TaskAssigned"
	EventTaskReassigned    = "TaskReassigned"
	EventTaskCompleted     = "TaskCompleted"
	EventTaskCancelled     = "TaskCancelled"
	EventTaskEscalated     = "TaskEscalated"
	EventFulfillmentFailed = "FulfillmentFailed"
	EventFulfillmentCancel = "FulfillmentCancelled"
	EventQCPassed          = "QCPassed"
	EventQCFailed          = "QCFailed"
)

// EventMeta is the common camelCase wire metadata for Kafka payloads.
type EventMeta struct {
	EventID       string
	EventType     string
	OccurredAt    time.Time
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	TraceID       string
}

// DomainEvent is the envelope for outbound integration events.
type DomainEvent struct {
	EventID       uuid.UUID
	Type          string
	OccurredAt    time.Time
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	TraceID       string
	Payload       map[string]any
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
	return nil
}

// Meta returns the wire EventMeta view of this event.
func (e DomainEvent) Meta() EventMeta {
	return EventMeta{
		EventID:       e.EventID.String(),
		EventType:     e.Type,
		OccurredAt:    e.OccurredAt,
		TenantID:      e.TenantID,
		WarehouseID:   e.WarehouseID,
		FulfillmentID: e.FulfillmentID,
		TraceID:       e.TraceID,
	}
}

// NewDomainEvent constructs a validated-ready envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, warehouseID, fulfillmentID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:       uuid.New(),
		Type:          eventType,
		OccurredAt:    time.Now().UTC(),
		TenantID:      tenantID,
		WarehouseID:   warehouseID,
		FulfillmentID: fulfillmentID,
		Payload:       payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventTaskAssigned, EventTaskReassigned, EventTaskCompleted,
		EventTaskCancelled, EventTaskEscalated:
		return TopicTaskEvents
	default:
		return TopicFulfillmentEvents
	}
}
