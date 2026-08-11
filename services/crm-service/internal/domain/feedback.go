package domain

import (
	"time"

	"github.com/google/uuid"
)

// Feedback kinds.
const (
	FeedbackCSAT = "csat"
	FeedbackNPS  = "nps"
	FeedbackCES  = "ces"
)

// Sentiment values.
const (
	SentimentPositive = "positive"
	SentimentNeutral  = "neutral"
	SentimentNegative = "negative"
)

// ValidSentiment reports whether s is a known sentiment.
func ValidSentiment(s string) bool {
	switch s {
	case SentimentPositive, SentimentNeutral, SentimentNegative:
		return true
	default:
		return false
	}
}

// Feedback is a generic survey response (NPS/CES).
type Feedback struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	TicketID       *uuid.UUID
	ConversationID *uuid.UUID
	Kind           string
	Score          int
	Comment        string
	CreatedAt      time.Time
}

// CSATResponse stores customer satisfaction scores.
type CSATResponse struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	TicketID       *uuid.UUID
	ConversationID *uuid.UUID
	Score          int
	Comment        string
	CreatedAt      time.Time
}

// AttachmentMeta is metadata for an uploaded file (blob stored elsewhere).
type AttachmentMeta struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	OwnerType   string
	OwnerID     uuid.UUID
	FileName    string
	ContentType string
	SizeBytes   int64
	URI         string
	CreatedAt   time.Time
}
