package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TicketStatus values.
const (
	TicketStatusOpen            = "open"
	TicketStatusPending         = "pending"
	TicketStatusInProgress      = "in_progress"
	TicketStatusWaitingCustomer = "waiting_customer"
	TicketStatusResolved        = "resolved"
	TicketStatusClosed          = "closed"
	TicketStatusReopened        = "reopened"
)

// Priority values.
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Category values.
const (
	CategoryOrder     = "order"
	CategoryPayment   = "payment"
	CategoryDelivery  = "delivery"
	CategoryRefund    = "refund"
	CategoryComplaint = "complaint"
	CategoryFraud     = "fraud"
	CategoryTechnical = "technical"
	CategoryLegal     = "legal"
	CategoryOther     = "other"
)

// TicketEventType values.
const (
	TicketEventCreated   = "created"
	TicketEventAssigned  = "assigned"
	TicketEventNote      = "note"
	TicketEventEscalated = "escalated"
	TicketEventResolved  = "resolved"
	TicketEventClosed    = "closed"
	TicketEventReopened  = "reopened"
	TicketEventMerged    = "merged"
	TicketEventStatus    = "status_changed"
)

var ticketTransitions = map[string]map[string]bool{
	TicketStatusOpen: {
		TicketStatusPending: true, TicketStatusInProgress: true,
		TicketStatusWaitingCustomer: true, TicketStatusResolved: true,
	},
	TicketStatusPending: {
		TicketStatusOpen: true, TicketStatusInProgress: true,
		TicketStatusWaitingCustomer: true, TicketStatusResolved: true,
	},
	TicketStatusInProgress: {
		TicketStatusPending: true, TicketStatusWaitingCustomer: true,
		TicketStatusResolved: true,
	},
	TicketStatusWaitingCustomer: {
		TicketStatusInProgress: true, TicketStatusPending: true,
		TicketStatusResolved: true,
	},
	TicketStatusResolved: {
		TicketStatusClosed: true, TicketStatusReopened: true,
		TicketStatusInProgress: true,
	},
	TicketStatusClosed: {
		TicketStatusReopened: true,
	},
	TicketStatusReopened: {
		TicketStatusPending: true, TicketStatusInProgress: true,
		TicketStatusWaitingCustomer: true, TicketStatusResolved: true,
	},
}

var priorityRank = map[string]int{
	PriorityLow: 1, PriorityNormal: 2, PriorityHigh: 3, PriorityUrgent: 4,
}

// ValidTicketStatus reports whether s is a known ticket status.
func ValidTicketStatus(s string) bool {
	_, ok := ticketTransitions[s]
	return ok
}

// ValidPriority reports whether p is a known priority.
func ValidPriority(p string) bool {
	_, ok := priorityRank[p]
	return ok
}

// ValidCategory reports whether c is a known category.
func ValidCategory(c string) bool {
	switch c {
	case CategoryOrder, CategoryPayment, CategoryDelivery, CategoryRefund,
		CategoryComplaint, CategoryFraud, CategoryTechnical, CategoryLegal, CategoryOther:
		return true
	default:
		return false
	}
}

// PriorityRank returns a numeric rank (higher = more urgent).
func PriorityRank(p string) int {
	return priorityRank[p]
}

// HigherPriority returns the higher of two priorities.
func HigherPriority(a, b string) string {
	if PriorityRank(a) >= PriorityRank(b) {
		return a
	}
	return b
}

// CanTransition reports whether from→to is allowed.
func CanTransition(from, to string) bool {
	m, ok := ticketTransitions[from]
	if !ok {
		return false
	}
	return m[to]
}

// Ticket is the support ticket aggregate.
type Ticket struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	CustomerID       uuid.UUID
	AssigneeID       *uuid.UUID
	TeamID           *uuid.UUID
	Status           string
	Priority         string
	Category         string
	Subject          string
	Description      string
	IdempotencyKey   string
	MergedIntoID     *uuid.UUID
	FirstResponseDue *time.Time
	ResolveDue       *time.Time
	FirstRespondedAt *time.Time
	ResolvedAt       *time.Time
	ClosedAt         *time.Time
	SLABreached      bool
	Tags             []string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Transition updates ticket status if allowed.
func (t *Ticket) Transition(to string, now time.Time) error {
	if !CanTransition(t.Status, to) {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, t.Status, to)
	}
	t.Status = to
	t.UpdatedAt = now
	switch to {
	case TicketStatusResolved:
		t.ResolvedAt = &now
	case TicketStatusClosed:
		t.ClosedAt = &now
	case TicketStatusReopened:
		t.ResolvedAt = nil
		t.ClosedAt = nil
	}
	return nil
}

// TicketEvent is an immutable ticket history row.
type TicketEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	TicketID  uuid.UUID
	ActorID   *uuid.UUID
	Type      string
	Payload   map[string]any
	CreatedAt time.Time
}

// TicketNote is an internal agent note.
type TicketNote struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	TicketID  uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	CreatedAt time.Time
}
