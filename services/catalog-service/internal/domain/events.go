package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topics under catalog.*.
const (
	EventProductCreated   = "ProductCreated"
	EventProductUpdated   = "ProductUpdated"
	EventProductPublished = "ProductPublished"
	EventProductArchived  = "ProductArchived"
	EventVariantCreated   = "VariantCreated"
	EventVariantUpdated   = "VariantUpdated"
	EventBundleCreated    = "BundleCreated"
	EventBundleUpdated    = "BundleUpdated"
	EventMediaAttached    = "MediaAttached"
	EventCategoryChanged  = "CategoryChanged"
	EventBrandChanged     = "BrandChanged"
	EventSupplierChanged  = "SupplierChanged"
	EventReindexProduct   = "ReindexProduct"
)

// Kafka topic names owned by this service.
const (
	TopicProductLifecycle = "catalog.product.lifecycle"
	TopicVariantEvents    = "catalog.variant.events"
	TopicBundleEvents     = "catalog.bundle.events"
	TopicMediaEvents      = "catalog.media.events"
	TopicCategoryEvents   = "catalog.category.events"
	TopicBrandEvents      = "catalog.brand.events"
	TopicSupplierEvents   = "catalog.supplier.events"
	TopicIndexCommands    = "catalog.index.commands"
)

// EventMeta is the common camelCase wire metadata for Kafka payloads.
type EventMeta struct {
	EventID    string
	EventType  string
	OccurredAt time.Time
	TenantID   uuid.UUID
	ProductID  uuid.UUID
	TraceID    string
}

// DomainEvent is the envelope for outbound integration events.
// Payloads are camelCase JSON at the wire layer.
type DomainEvent struct {
	EventID    uuid.UUID
	Type       string
	OccurredAt time.Time
	TenantID   uuid.UUID
	ProductID  uuid.UUID
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
	return nil
}

// Meta returns the wire EventMeta view of this event.
func (e DomainEvent) Meta() EventMeta {
	return EventMeta{
		EventID:    e.EventID.String(),
		EventType:  e.Type,
		OccurredAt: e.OccurredAt,
		TenantID:   e.TenantID,
		ProductID:  e.ProductID,
		TraceID:    e.TraceID,
	}
}

// NewDomainEvent constructs a validated-ready envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, productID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:    uuid.New(),
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		TenantID:   tenantID,
		ProductID:  productID,
		Payload:    payload,
	}
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventProductCreated, EventProductUpdated, EventProductPublished, EventProductArchived:
		return TopicProductLifecycle
	case EventVariantCreated, EventVariantUpdated:
		return TopicVariantEvents
	case EventBundleCreated, EventBundleUpdated:
		return TopicBundleEvents
	case EventMediaAttached:
		return TopicMediaEvents
	case EventCategoryChanged:
		return TopicCategoryEvents
	case EventBrandChanged:
		return TopicBrandEvents
	case EventSupplierChanged:
		return TopicSupplierEvents
	case EventReindexProduct:
		return TopicIndexCommands
	default:
		return TopicProductLifecycle
	}
}
