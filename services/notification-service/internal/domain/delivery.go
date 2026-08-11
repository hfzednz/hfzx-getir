package domain

import (
	"time"

	"github.com/google/uuid"
)

// DeliveryAttempt records one provider dispatch try.
type DeliveryAttempt struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	MessageID   uuid.UUID
	AttemptNo   int
	Provider    string
	Status      string // success|failed
	ProviderRef string
	Error       string
	CreatedAt   time.Time
}

// DeliveryEvent is an analytics / receipt event (opened, clicked, bounced, …).
type DeliveryEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	MessageID uuid.UUID
	Type      string
	Payload   map[string]any
	CreatedAt time.Time
}

// DLQItem is a dead-letter entry after max retries.
type DLQItem struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	MessageID uuid.UUID
	Reason    string
	Payload   map[string]any
	CreatedAt time.Time
}

// ProviderRoute maps channel → provider name.
type ProviderRoute struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Channel   Channel
	Provider  string
	Priority  int
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
