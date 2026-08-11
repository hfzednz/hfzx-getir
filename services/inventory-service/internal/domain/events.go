package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topics under inventory.*.
const (
	EventInventoryCreated     = "InventoryCreated"
	EventInventoryUpdated     = "InventoryUpdated"
	EventStockAdjusted        = "StockAdjusted"
	EventStockReceived        = "StockReceived"
	EventStockExpired         = "StockExpired"
	EventStockReserved        = "StockReserved"
	EventReservationReleased  = "ReservationReleased"
	EventReservationExpired   = "ReservationExpired"
	EventReservationConfirmed = "ReservationConfirmed"
	EventReservationConsumed  = "ReservationConsumed"
	EventStockTransferred     = "StockTransferred"
	EventStockCountCompleted  = "StockCountCompleted"
	EventReindexStock         = "ReindexStock"
)

// Kafka topic names owned by this service.
const (
	TopicStockEvents       = "inventory.stock.events"
	TopicReservationEvents = "inventory.reservation.events"
	TopicTransferEvents    = "inventory.transfer.events"
	TopicCountEvents       = "inventory.count.events"
	TopicIndexCommands     = "inventory.index.commands"
)

// EventMeta is the common camelCase wire metadata for Kafka payloads.
type EventMeta struct {
	EventID     string
	EventType   string
	OccurredAt  time.Time
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	TraceID     string
}

// DomainEvent is the envelope for outbound integration events.
// Payloads are camelCase JSON at the wire layer.
type DomainEvent struct {
	EventID     uuid.UUID
	Type        string
	OccurredAt  time.Time
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
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
	return nil
}

// Meta returns the wire EventMeta view of this event.
func (e DomainEvent) Meta() EventMeta {
	return EventMeta{
		EventID:     e.EventID.String(),
		EventType:   e.Type,
		OccurredAt:  e.OccurredAt,
		TenantID:    e.TenantID,
		WarehouseID: e.WarehouseID,
		VariantID:   e.VariantID,
		TraceID:     e.TraceID,
	}
}

// NewDomainEvent constructs a validated-ready envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, warehouseID, variantID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:     uuid.New(),
		Type:        eventType,
		OccurredAt:  time.Now().UTC(),
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		VariantID:   variantID,
		Payload:     payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventInventoryCreated, EventInventoryUpdated, EventStockAdjusted,
		EventStockReceived, EventStockExpired:
		return TopicStockEvents
	case EventStockReserved, EventReservationReleased, EventReservationExpired,
		EventReservationConfirmed, EventReservationConsumed:
		return TopicReservationEvents
	case EventStockTransferred:
		return TopicTransferEvents
	case EventStockCountCompleted:
		return TopicCountEvents
	case EventReindexStock:
		return TopicIndexCommands
	default:
		return TopicStockEvents
	}
}
