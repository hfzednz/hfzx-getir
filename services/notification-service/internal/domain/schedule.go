package domain

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleStatus is the schedule lifecycle.
type ScheduleStatus string

const (
	SchedulePending   ScheduleStatus = "pending"
	ScheduleProcessed ScheduleStatus = "processed"
	ScheduleCancelled ScheduleStatus = "cancelled"
)

// Schedule is a delayed send.
type Schedule struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	Channel        Channel
	Priority       Priority
	TemplateKey    string
	Locale         string
	Recipient      string
	Subject        string
	Body           string
	Vars           map[string]string
	IdempotencyKey string
	SendAt         time.Time
	Status         ScheduleStatus
	MessageID      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
