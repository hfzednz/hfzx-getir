package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicLiveOpsEvents    = "liveops.events"
)

const (
	EventFeatureFlagCreated    = "FeatureFlagCreated"
	EventFeatureFlagUpdated    = "FeatureFlagUpdated"
	EventFeatureEnabled        = "FeatureEnabled"
	EventFeatureDisabled       = "FeatureDisabled"
	EventExperimentStarted     = "ExperimentStarted"
	EventExperimentCompleted   = "ExperimentCompleted"
	EventConfigurationUpdated  = "ConfigurationUpdated"
	EventRollbackExecuted      = "RollbackExecuted"
)

func TopicForEvent(string) string { return TopicLiveOpsEvents }

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
