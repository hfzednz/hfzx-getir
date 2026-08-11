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
	EventSecurityAlertCreated      = "SecurityAlertCreated"
	EventThreatDetected            = "ThreatDetected"
	EventPolicyViolated            = "PolicyViolated"
	EventSecretRotated             = "SecretRotated"
	EventCertificateRenewed        = "CertificateRenewed"
	EventIncidentOpened            = "IncidentOpened"
	EventIncidentClosed            = "IncidentClosed"
	EventComplianceAuditCompleted  = "ComplianceAuditCompleted"
)

const TopicSecurityEvents = "security.events"

func TopicForEvent(string) string { return TopicSecurityEvents }

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
