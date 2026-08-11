package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants for Kafka topics under customer.*.
const (
	EventCustomerCreated   = "CustomerCreated"
	EventCustomerUpdated   = "CustomerUpdated"
	EventProfileDeleted    = "ProfileDeleted"
	EventAddressAdded      = "AddressAdded"
	EventAddressUpdated    = "AddressUpdated"
	EventAddressRemoved    = "AddressRemoved"
	EventPreferenceChanged = "PreferenceChanged"
	EventAvatarUpdated     = "AvatarUpdated"
	EventConsentChanged    = "ConsentChanged"
	EventSegmentChanged    = "SegmentChanged"
	EventExportRequested   = "ExportRequested"
	EventDeletionRequested = "DeletionRequested"
)

// Kafka topic names owned by this service.
const (
	TopicProfileLifecycle = "customer.profile.lifecycle"
	TopicAddressEvents    = "customer.address.events"
	TopicPreferenceEvents = "customer.preference.events"
	TopicMediaEvents      = "customer.media.events"
	TopicConsentEvents    = "customer.consent.events"
	TopicSegmentEvents    = "customer.segment.events"
	TopicPrivacyEvents    = "customer.privacy.events"
)

// EventMeta is the common camelCase wire metadata for Kafka payloads.
type EventMeta struct {
	EventID     string
	EventType   string
	OccurredAt  time.Time
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	ProfileID   uuid.UUID
	TraceID     string
}

// DomainEvent is the envelope for outbound integration events.
// Payloads are camelCase JSON at the wire layer.
type DomainEvent struct {
	EventID     uuid.UUID
	Type        string
	OccurredAt  time.Time
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	ProfileID   uuid.UUID
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
	if e.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
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
		PrincipalID: e.PrincipalID,
		ProfileID:   e.ProfileID,
		TraceID:     e.TraceID,
	}
}

// NewDomainEvent constructs a validated-ready envelope with a fresh event id.
func NewDomainEvent(eventType string, tenantID, principalID, profileID uuid.UUID, payload map[string]any) DomainEvent {
	if payload == nil {
		payload = map[string]any{}
	}
	return DomainEvent{
		EventID:     uuid.New(),
		Type:        eventType,
		OccurredAt:  time.Now().UTC(),
		TenantID:    tenantID,
		PrincipalID: principalID,
		ProfileID:   profileID,
		Payload:     payload,
	}
}
