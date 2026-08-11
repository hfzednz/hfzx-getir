package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain event type constants.
const (
	EventJournalPosted     = "JournalPosted"
	EventInvoiceGenerated  = "InvoiceGenerated"
	EventCreditNoteIssued  = "CreditNoteIssued"
	EventAccountEnsured    = "AccountEnsured"
)

// Kafka topic names owned by this service.
const (
	TopicLedgerLifecycle = "ledger.lifecycle"
)

// DomainEvent is the envelope for outbound integration events.
type DomainEvent struct {
	EventID    uuid.UUID
	Type       string
	OccurredAt time.Time
	TenantID   uuid.UUID
	EntityID   uuid.UUID
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
	return nil
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicLedgerLifecycle
}

// LedgerEvent is the append-only timeline row projection.
type LedgerEvent struct {
	ID         uuid.UUID
	EntityID   uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	ActorType  string
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e LedgerEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: event id required", ErrInvalidArgument)
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
