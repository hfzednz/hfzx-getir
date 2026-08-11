package domain

import (
	"time"

	"github.com/google/uuid"
)

// SLAPolicy defines first-response and resolve targets per priority.
type SLAPolicy struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	Name                 string
	Priority             string
	FirstResponseMinutes int
	ResolveMinutes       int
	Active               bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ComputeDueTimes returns FirstResponseDue and ResolveDue from policy start.
func (p SLAPolicy) ComputeDueTimes(start time.Time) (firstResponse, resolve time.Time) {
	firstResponse = start.Add(time.Duration(p.FirstResponseMinutes) * time.Minute)
	resolve = start.Add(time.Duration(p.ResolveMinutes) * time.Minute)
	return firstResponse, resolve
}

// Escalation records a ticket escalation.
type Escalation struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	TicketID       uuid.UUID
	FromPriority   string
	ToPriority     string
	Reason         string
	TriggeredBySLA bool
	CreatedAt      time.Time
}

// NextPriority bumps priority one level (capped at urgent).
func NextPriority(p string) string {
	switch p {
	case PriorityLow:
		return PriorityNormal
	case PriorityNormal:
		return PriorityHigh
	case PriorityHigh, PriorityUrgent:
		return PriorityUrgent
	default:
		return PriorityHigh
	}
}
