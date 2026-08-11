package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventWalletCredited = "WalletCredited"
	EventWalletDebited  = "WalletDebited"
	EventWalletHeld     = "WalletHeld"
	EventWalletReleased = "WalletReleased"
	EventWalletTransfer = "WalletTransferred"
	EventWalletAdjusted = "WalletAdjusted"
)

const TopicWalletLifecycle = "wallet.lifecycle"

// OutboxStatus is the transactional outbox row lifecycle.
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
)

// Valid reports whether the outbox status is recognized.
func (s OutboxStatus) Valid() bool {
	switch s {
	case OutboxStatusPending, OutboxStatusPublished, OutboxStatusFailed:
		return true
	default:
		return false
	}
}

// OutboxMessage is a transactional outbox row awaiting publish.
type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WalletID    uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      OutboxStatus
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

// Validate checks outbox message invariants.
func (m OutboxMessage) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: outbox id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if m.Topic == "" {
		return fmt.Errorf("%w: topic required", ErrInvalidArgument)
	}
	if !m.Status.Valid() {
		return fmt.Errorf("%w: invalid outbox status %q", ErrInvalidArgument, m.Status)
	}
	return nil
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(string) string {
	return TopicWalletLifecycle
}

// AuditEntry records an admin money mutation.
type AuditEntry struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WalletID    uuid.UUID
	Action      string
	ActorID     *uuid.UUID
	AmountMinor int64
	Currency    string
	Detail      map[string]any
	CreatedAt   time.Time
}
