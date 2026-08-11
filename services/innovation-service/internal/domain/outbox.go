package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicInnovationEvents = "innovation.events"
)

const (
	EventSimulationStarted         = "SimulationStarted"
	EventSimulationCompleted       = "SimulationCompleted"
	EventInnovationEnabled         = "InnovationEnabled"
	EventResearchExperimentCreated = "ResearchExperimentCreated"
	EventEdgeNodeRegistered        = "EdgeNodeRegistered"
	EventIoTDeviceConnected        = "IoTDeviceConnected"
	EventRobotAssigned             = "RobotAssigned"
	EventDroneMissionCreated       = "DroneMissionCreated"
)

func TopicForEvent(string) string { return TopicInnovationEvents }

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
