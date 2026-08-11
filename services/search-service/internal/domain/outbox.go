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
	EventSearchPerformed      = "SearchPerformed"
	EventSuggestionClicked    = "SuggestionClicked"
	EventProductRankUpdated   = "ProductRankUpdated"
	EventEmbeddingGenerated   = "EmbeddingGenerated"
	EventIndexUpdated         = "IndexUpdated"
	EventTrendingUpdated      = "TrendingUpdated"
	EventRecommendationProxy  = "RecommendationShown" // when search embeds rails
)

const TopicSearchEvents = "search.events"

func TopicForEvent(string) string { return TopicSearchEvents }

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
