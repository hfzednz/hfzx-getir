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

// Outbound review/trust event types.
const (
	EventReviewCreated      = "ReviewCreated"
	EventReviewUpdated      = "ReviewUpdated"
	EventReviewDeleted      = "ReviewDeleted"
	EventRatingSubmitted    = "RatingSubmitted"
	EventMediaAttached      = "MediaAttached"
	EventReviewReported     = "ReviewReported"
	EventReviewApproved     = "ReviewApproved"
	EventReviewRejected     = "ReviewRejected"
	EventTrustScoreUpdated  = "TrustScoreUpdated"
	EventReputationUpdated  = "ReputationUpdated"
	EventModerationQueued   = "ModerationQueued"
	EventQualityScored      = "QualityScored"
)

// TopicReviewEvents is the primary outbound topic.
const TopicReviewEvents = "review.events"

// TopicForEvent maps event type to Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicReviewEvents
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
