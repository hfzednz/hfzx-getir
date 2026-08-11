package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicPlatformEvents   = "platform.events"
)

const (
	EventDeploymentStarted   = "DeploymentStarted"
	EventDeploymentCompleted = "DeploymentCompleted"
	EventRollbackTriggered   = "RollbackTriggered"
	EventScalingTriggered    = "ScalingTriggered"
	EventBackupCompleted     = "BackupCompleted"
	EventRecoveryStarted     = "RecoveryStarted"
	EventAlertTriggered      = "AlertTriggered"
)

func TopicForEvent(string) string { return TopicPlatformEvents }

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

type Deployment struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Service     string
	Environment string
	Strategy    string // rolling|blue_green|canary|shadow
	ImageTag    string
	Status      string // started|succeeded|failed|rolled_back
	StartedAt   time.Time
	CompletedAt *time.Time
}

type ScalingEvent struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Service     string
	Environment string
	FromReplicas int
	ToReplicas   int
	Reason      string
	CreatedAt   time.Time
}

type BackupJob struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Kind        string // postgres|volume|object
	Target      string
	Status      string // running|completed|failed
	Location    string
	StartedAt   time.Time
	CompletedAt *time.Time
}

type RecoveryJob struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Kind        string // region_failover|restore|chaos_drill
	Status      string // started|completed|failed
	Notes       string
	StartedAt   time.Time
	CompletedAt *time.Time
}

type AlertEvent struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Severity   string
	Status     string // firing|resolved
	Labels     map[string]string
	FiredAt    time.Time
	ResolvedAt *time.Time
}

type CostSnapshot struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Environment string
	AmountMinor int64 // USD cents
	Currency    string
	Period      string
	CreatedAt   time.Time
}

type SLOReport struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Service     string
	Objective   float64
	Actual      float64
	BudgetLeft  float64
	Window      string
	CreatedAt   time.Time
}

func ValidStrategy(s string) bool {
	switch s {
	case "rolling", "blue_green", "canary", "shadow":
		return true
	default:
		return false
	}
}

func BurnRate(objective, actual float64) float64 {
	if objective <= 0 {
		return 0
	}
	budget := 1 - objective/100
	if budget <= 0 {
		return 0
	}
	consumed := (objective - actual) / 100
	if consumed < 0 {
		return 0
	}
	return consumed / budget
}
