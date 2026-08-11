package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

const (
	EventRecommendationShown   = "RecommendationShown"
	EventRecommendationClicked = "RecommendationClicked"
	EventSignalIngested        = "SignalIngested"
)

const TopicRecEvents = "recommendation.events"

func TopicForEvent(string) string { return TopicRecEvents }

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
