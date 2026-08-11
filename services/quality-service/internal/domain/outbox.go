package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicQualityEvents    = "quality.events"
)

const (
	EventTestStarted          = "TestStarted"
	EventTestCompleted        = "TestCompleted"
	EventCoverageGenerated    = "CoverageGenerated"
	EventQualityGatePassed    = "QualityGatePassed"
	EventQualityGateFailed    = "QualityGateFailed"
	EventCertificationIssued  = "CertificationIssued"
	EventRegressionCompleted  = "RegressionCompleted"
)

func TopicForEvent(string) string { return TopicQualityEvents }

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
