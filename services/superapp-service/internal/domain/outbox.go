package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicSuperAppEvents   = "superapp.events"
)

const (
	EventPluginInstalled   = "PluginInstalled"
	EventPluginRemoved     = "PluginRemoved"
	EventPluginUpdated     = "PluginUpdated"
	EventMiniAppLaunched   = "MiniAppLaunched"
	EventWidgetAdded       = "WidgetAdded"
	EventModuleActivated   = "ModuleActivated"
	EventPermissionGranted = "PermissionGranted"
)

func TopicForEvent(string) string { return TopicSuperAppEvents }

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
