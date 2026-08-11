package domain

import (
	"time"

	"github.com/google/uuid"
)

// Case types.
const (
	CaseTypeRefund    = "refund"
	CaseTypeComplaint = "complaint"
	CaseTypeCourier   = "courier"
	CaseTypeWarehouse = "warehouse"
	CaseTypeFraud     = "fraud"
	CaseTypePayment   = "payment"
	CaseTypeTechnical = "technical"
	CaseTypeLegal     = "legal"
)

// Case statuses.
const (
	CaseStatusOpen       = "open"
	CaseStatusInProgress = "in_progress"
	CaseStatusResolved   = "resolved"
	CaseStatusClosed     = "closed"
)

// ValidCaseType reports whether t is a known case type.
func ValidCaseType(t string) bool {
	switch t {
	case CaseTypeRefund, CaseTypeComplaint, CaseTypeCourier, CaseTypeWarehouse,
		CaseTypeFraud, CaseTypePayment, CaseTypeTechnical, CaseTypeLegal:
		return true
	default:
		return false
	}
}

// Case is an investigation / ops case linked to a customer or ticket.
type Case struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	TicketID   *uuid.UUID
	Type       string
	Status     string
	Title      string
	Details    string
	AssigneeID *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
