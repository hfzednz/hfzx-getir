package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventSettlementCreated   = "SettlementCreated"
	EventSettlementSubmitted = "SettlementSubmitted"
	EventSettlementApproved  = "SettlementApproved"
	EventSettlementCompleted = "SettlementCompleted"
	EventSettlementFailed    = "SettlementFailed"
	EventSettlementReconciled = "SettlementReconciled"
)

const (
	TopicSettlementLifecycle = "settlement.lifecycle"
)

// SettlementEvent is the append-only timeline row.
type SettlementEvent struct {
	ID         uuid.UUID
	BatchID    uuid.UUID
	TenantID   uuid.UUID
	Type       string
	Payload    map[string]any
	ActorID    *uuid.UUID
	ActorType  string
	OccurredAt time.Time
	CreatedAt  time.Time
}

// Validate checks timeline event invariants.
func (e SettlementEvent) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: event id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: event type required", ErrInvalidArgument)
	}
	return nil
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicSettlementLifecycle
}
