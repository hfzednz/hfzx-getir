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
	EventEventIngested      = "EventIngested"
	EventSchemaRegistered   = "SchemaRegistered"
	EventAggregateUpdated   = "AggregateUpdated"
	EventMartRefreshed      = "MartRefreshed"
	EventReportGenerated    = "ReportGenerated"
	EventExperimentDecided  = "ExperimentDecided"
	EventAlertFired         = "AlertFired"
	EventQualityFailed      = "QualityFailed"
)

const TopicDataEvents = "data.events"

func TopicForEvent(string) string { return TopicDataEvents }

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
