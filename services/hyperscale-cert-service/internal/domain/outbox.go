package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicHyperscaleEvents = "hyperscale.events"
)

const (
	EventAuditCompleted           = "AuditCompleted"
	EventBenchmarkRecorded        = "BenchmarkRecorded"
	EventChaosExperimentCompleted = "ChaosExperimentCompleted"
	EventOptimizationApplied      = "OptimizationApplied"
	EventHyperscaleCertified      = "HyperscaleCertified"
)

func TopicForEvent(string) string { return TopicHyperscaleEvents }

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
