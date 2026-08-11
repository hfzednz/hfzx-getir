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
	EventModelTrained         = "ModelTrained"
	EventModelDeployed        = "ModelDeployed"
	EventInferenceCompleted   = "InferenceCompleted"
	EventPredictionGenerated  = "PredictionGenerated"
	EventFeatureUpdated       = "FeatureUpdated"
	EventAgentExecuted        = "AgentExecuted"
	EventAutomationTriggered  = "AutomationTriggered"
	EventDriftDetected        = "DriftDetected"
)

const TopicAIEvents = "ai.events"

func TopicForEvent(string) string { return TopicAIEvents }

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
