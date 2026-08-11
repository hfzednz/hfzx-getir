package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// QCResult is the inspection outcome.
type QCResult string

const (
	QCResultPending QCResult = "pending"
	QCResultPass    QCResult = "passed"
	QCResultFail    QCResult = "failed"
	QCResultWaived  QCResult = "waived"

	// Aliases.
	QCPending = QCResultPending
	QCPassed  = QCResultPass
	QCFailed  = QCResultFail
	QCWaived  = QCResultWaived
)

func (r QCResult) Valid() bool {
	switch r {
	case QCResultPending, QCResultPass, QCResultFail, QCResultWaived:
		return true
	default:
		return false
	}
}

func (r QCResult) IsTerminal() bool {
	return r == QCResultPass || r == QCResultFail || r == QCResultWaived
}

// QCInspection is a quality check on a fulfill unit / package.
type QCInspection struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	FulfillmentID  uuid.UUID
	TaskID         *uuid.UUID
	StationID      *uuid.UUID
	DispatchUnitID *uuid.UUID
	InspectorID    *uuid.UUID
	Result         QCResult
	Checklist      []map[string]any
	Notes          string
	DefectCodes    []string
	Defects        []map[string]any
	InspectedAt    *time.Time
	CompletedAt    *time.Time
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks QC inspection invariants.
func (q QCInspection) Validate() error {
	if q.ID == uuid.Nil {
		return fmt.Errorf("%w: qc inspection id required", ErrInvalidArgument)
	}
	if q.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if q.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !q.Result.Valid() {
		return fmt.Errorf("%w: invalid qc result %q", ErrInvalidArgument, q.Result)
	}
	return nil
}

// Pass marks the inspection passed.
func (q *QCInspection) Pass(inspectorID *uuid.UUID, notes string) error {
	if q.Result.IsTerminal() {
		return fmt.Errorf("%w: result %s", ErrAlreadyTerminal, q.Result)
	}
	now := time.Now().UTC()
	q.Result = QCResultPass
	q.InspectorID = inspectorID
	q.Notes = notes
	q.InspectedAt = &now
	q.CompletedAt = &now
	q.UpdatedAt = now
	return nil
}

// Fail marks the inspection failed with optional defect codes.
func (q *QCInspection) Fail(inspectorID *uuid.UUID, defectCodes []string, notes string) error {
	if q.Result.IsTerminal() {
		return fmt.Errorf("%w: result %s", ErrAlreadyTerminal, q.Result)
	}
	now := time.Now().UTC()
	q.Result = QCResultFail
	q.InspectorID = inspectorID
	q.DefectCodes = defectCodes
	q.Notes = notes
	q.InspectedAt = &now
	q.CompletedAt = &now
	q.UpdatedAt = now
	return nil
}

// Waive marks the inspection waived (supervisor override).
func (q *QCInspection) Waive(inspectorID *uuid.UUID, notes string) error {
	if q.Result.IsTerminal() {
		return fmt.Errorf("%w: result %s", ErrAlreadyTerminal, q.Result)
	}
	now := time.Now().UTC()
	q.Result = QCResultWaived
	q.InspectorID = inspectorID
	q.Notes = notes
	q.InspectedAt = &now
	q.CompletedAt = &now
	q.UpdatedAt = now
	return nil
}
