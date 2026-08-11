package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicEnterpriseEvents = "enterprise.events"
)

const (
	EventPolicyApproved           = "PolicyApproved"
	EventProjectCreated           = "ProjectCreated"
	EventRiskIdentified           = "RiskIdentified"
	EventAuditCompleted           = "AuditCompleted"
	EventMeetingScheduled         = "MeetingScheduled"
	EventDecisionRecorded         = "DecisionRecorded"
	EventContinuityPlanActivated  = "ContinuityPlanActivated"
)

func TopicForEvent(string) string { return TopicEnterpriseEvents }

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
