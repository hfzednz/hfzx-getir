package domain

import (
	"time"

	"github.com/google/uuid"
)

// Conversation statuses.
const (
	ConversationStatusActive    = "active"
	ConversationStatusTransferred = "transferred"
	ConversationStatusEnded     = "ended"
)

// Message sender roles.
const (
	SenderCustomer = "customer"
	SenderAgent    = "agent"
	SenderAI       = "ai"
	SenderSystem   = "system"
)

// ValidSender reports whether s is a known sender role.
func ValidSender(s string) bool {
	switch s {
	case SenderCustomer, SenderAgent, SenderAI, SenderSystem:
		return true
	default:
		return false
	}
}

// Conversation is a live-chat session.
type Conversation struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CustomerID   uuid.UUID
	AgentID      *uuid.UUID
	TicketID     *uuid.UUID
	Status       string
	Channel      string
	TransferredFrom *uuid.UUID
	StartedAt    time.Time
	EndedAt      *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Message is a chat message within a conversation.
type Message struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	ConversationID uuid.UUID
	SenderRole     string
	SenderID       *uuid.UUID
	Body           string
	Sentiment      string
	CreatedAt      time.Time
}
