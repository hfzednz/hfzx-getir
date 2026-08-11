package domain

import (
	"time"

	"github.com/google/uuid"
)

// Agent statuses.
const (
	AgentStatusAvailable = "available"
	AgentStatusBusy      = "busy"
	AgentStatusOffline   = "offline"
)

// Agent is a support workforce member.
type Agent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	DisplayName string
	TeamID    *uuid.UUID
	Status    string
	SkillIDs  []uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Team groups agents.
type Team struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Skill is a routing capability tag.
type Skill struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	CreatedAt time.Time
}
