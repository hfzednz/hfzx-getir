package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicAutonomyEvents   = "autonomy.events"
)

const (
	EventAutonomyAuditCompleted = "AutonomyAuditCompleted"
	EventSelfHealExecuted       = "SelfHealExecuted"
	EventAICTOReviewCompleted   = "AICTOReviewCompleted"
	EventEvolutionTaskCreated   = "EvolutionTaskCreated"
	EventAutonomousReleaseScored = "AutonomousReleaseScored"
	EventGenesisCertified       = "GenesisCertified"
)

func TopicForEvent(string) string { return TopicAutonomyEvents }

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
