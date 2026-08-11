package domain

import (
	"time"

	"github.com/google/uuid"
)

// Outbox statuses.
const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

// Outbound CRM event types.
const (
	EventTicketCreated     = "TicketCreated"
	EventTicketAssigned    = "TicketAssigned"
	EventTicketEscalated   = "TicketEscalated"
	EventTicketResolved    = "TicketResolved"
	EventTicketClosed      = "TicketClosed"
	EventTicketReopened    = "TicketReopened"
	EventChatStarted       = "ChatStarted"
	EventChatEnded         = "ChatEnded"
	EventComplaintCreated  = "ComplaintCreated"
	EventRefundRequested   = "RefundRequested"
	EventFeedbackReceived  = "FeedbackReceived"
	EventCSATCompleted     = "CSATCompleted"
	EventSLABreached       = "SLABreached"
	EventArticlePublished  = "ArticlePublished"
)

// TopicCRMEvents is the primary outbound topic.
const TopicCRMEvents = "crm.events"

// TopicForEvent maps event type to Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicCRMEvents
}

// OutboxMessage is a transactional outbox row.
type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AggregateID uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}
