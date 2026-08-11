package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageStatus is the message lifecycle.
type MessageStatus string

const (
	MessageQueued     MessageStatus = "queued"
	MessageSending    MessageStatus = "sending"
	MessageSent       MessageStatus = "sent"
	MessageFailed     MessageStatus = "failed"
	MessageSuppressed MessageStatus = "suppressed"
)

// Message is a single notification send unit.
type Message struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	OrderID        *uuid.UUID // opaque optional
	Channel        Channel
	Priority       Priority
	TemplateKey    string
	TemplateID     *uuid.UUID
	Locale         string
	Subject        string
	Body           string
	Recipient      string // email / phone / opaque address
	Status         MessageStatus
	IdempotencyKey string
	Vars           map[string]string
	SuppressReason string
	Attempts       int
	MaxAttempts    int
	LastError      string
	Provider       string
	ProviderRef    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	SentAt         *time.Time
}

// DefaultMaxAttempts is the retry budget before DLQ.
const DefaultMaxAttempts = 3
