package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicOpenPlatformEvents = "openplatform.events"
)

const (
	EventApiKeyCreated      = "ApiKeyCreated"
	EventWebhookRegistered  = "WebhookRegistered"
	EventWebhookDelivered   = "WebhookDelivered"
	EventSdkGenerated       = "SdkGenerated"
	EventApiVersionReleased = "ApiVersionReleased"
	EventPartnerIntegrated  = "PartnerIntegrated"
)

func TopicForEvent(string) string { return TopicOpenPlatformEvents }

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
