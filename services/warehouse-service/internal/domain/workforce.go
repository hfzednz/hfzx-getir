package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmployeeRole is the WH-scoped work role.
type EmployeeRole string

const (
	EmployeeRolePicker      EmployeeRole = "picker"
	EmployeeRolePacker      EmployeeRole = "packer"
	EmployeeRoleDispatcher  EmployeeRole = "dispatcher"
	EmployeeRoleQC          EmployeeRole = "qc"
	EmployeeRoleSupervisor  EmployeeRole = "supervisor"
	EmployeeRoleRunner      EmployeeRole = "runner"
	EmployeeRoleMaintenance EmployeeRole = "maintenance"

	// Short aliases.
	RolePicker      = EmployeeRolePicker
	RolePacker      = EmployeeRolePacker
	RoleDispatcher  = EmployeeRoleDispatcher
	RoleQC          = EmployeeRoleQC
	RoleSupervisor  = EmployeeRoleSupervisor
	RoleRunner      = EmployeeRoleRunner
	RoleMaintenance = EmployeeRoleMaintenance
)

func (r EmployeeRole) Valid() bool {
	switch r {
	case EmployeeRolePicker, EmployeeRolePacker, EmployeeRoleDispatcher, EmployeeRoleQC,
		EmployeeRoleSupervisor, EmployeeRoleRunner, EmployeeRoleMaintenance:
		return true
	default:
		return false
	}
}

// EmployeeStatus is roster status.
type EmployeeStatus string

const (
	EmployeeStatusActive    EmployeeStatus = "active"
	EmployeeStatusInactive  EmployeeStatus = "inactive"
	EmployeeStatusSuspended EmployeeStatus = "suspended"
)

func (s EmployeeStatus) Valid() bool {
	switch s {
	case EmployeeStatusActive, EmployeeStatusInactive, EmployeeStatusSuspended:
		return true
	default:
		return false
	}
}

// ShiftStatus is shift lifecycle.
type ShiftStatus string

const (
	ShiftStatusScheduled  ShiftStatus = "scheduled"
	ShiftStatusClockedIn  ShiftStatus = "clocked_in"
	ShiftStatusOnBreak    ShiftStatus = "on_break"
	ShiftStatusClockedOut ShiftStatus = "clocked_out"
	ShiftStatusCompleted  ShiftStatus = "completed"
	ShiftStatusCancelled  ShiftStatus = "cancelled"
	ShiftStatusNoShow     ShiftStatus = "no_show"
	ShiftStatusActive     ShiftStatus = "active" // alias for clocked_in in SQL mapping
)

func (s ShiftStatus) Valid() bool {
	switch s {
	case ShiftStatusScheduled, ShiftStatusClockedIn, ShiftStatusOnBreak,
		ShiftStatusClockedOut, ShiftStatusCompleted, ShiftStatusCancelled,
		ShiftStatusNoShow, ShiftStatusActive:
		return true
	default:
		return false
	}
}

// AttendanceEventType marks clock/break events.
type AttendanceEventType string

const (
	AttendanceClockIn    AttendanceEventType = "clock_in"
	AttendanceClockOut   AttendanceEventType = "clock_out"
	AttendanceBreakStart AttendanceEventType = "break_start"
	AttendanceBreakEnd   AttendanceEventType = "break_end"
)

func (t AttendanceEventType) Valid() bool {
	switch t {
	case AttendanceClockIn, AttendanceClockOut, AttendanceBreakStart, AttendanceBreakEnd:
		return true
	default:
		return false
	}
}

// BreakInterval is an open/closed break within a shift.
type BreakInterval struct {
	StartedAt time.Time
	EndedAt   *time.Time
	Reason    string
}

// Employee is WH-scoped workforce (not HR payroll).
type Employee struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	PrincipalID *uuid.UUID // opaque identity principal
	ExternalRef string
	BadgeCode   string
	DisplayName string
	Role        EmployeeRole
	Status      EmployeeStatus
	Active      bool
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Shift is a scheduled / active work window.
type Shift struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	WarehouseID  uuid.UUID
	EmployeeID   uuid.UUID
	Status       ShiftStatus
	Role         EmployeeRole
	PlannedStart time.Time
	PlannedEnd   time.Time
	ActualStart  *time.Time
	ActualEnd    *time.Time
	ClockInAt    time.Time
	ClockOutAt   *time.Time
	StationID    *uuid.UUID
	Breaks       []BreakInterval
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AttendanceEvent is a clock/break marker.
type AttendanceEvent struct {
	ID          uuid.UUID
	EmployeeID  uuid.UUID
	ShiftID     *uuid.UUID
	WarehouseID uuid.UUID
	EventType   AttendanceEventType
	OccurredAt  time.Time
	Source      string
	Metadata    map[string]any
	CreatedAt   time.Time
}

// BreakPeriod is an open/closed break interval (SQL projection).
type BreakPeriod struct {
	ID         uuid.UUID
	ShiftID    uuid.UUID
	EmployeeID uuid.UUID
	StartedAt  time.Time
	EndedAt    *time.Time
	BreakKind  string
	Metadata   map[string]any
	CreatedAt  time.Time
}

// Validate checks employee invariants.
func (e Employee) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: employee id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if e.DisplayName == "" {
		return fmt.Errorf("%w: display_name required", ErrInvalidArgument)
	}
	if e.Role != "" && !e.Role.Valid() {
		return fmt.Errorf("%w: invalid employee role %q", ErrInvalidArgument, e.Role)
	}
	return nil
}

// IsActive reports whether the employee can be assigned work.
func (e Employee) IsActive() bool {
	if e.DeletedAt != nil {
		return false
	}
	if e.Status != "" {
		return e.Status == EmployeeStatusActive
	}
	return e.Active
}

// Validate checks shift invariants.
func (s Shift) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: shift id required", ErrInvalidArgument)
	}
	if s.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if s.EmployeeID == uuid.Nil {
		return fmt.Errorf("%w: employee_id required", ErrInvalidArgument)
	}
	if s.Status != "" && !s.Status.Valid() {
		return fmt.Errorf("%w: invalid shift status %q", ErrInvalidArgument, s.Status)
	}
	return nil
}

// ClockIn starts an active shift.
func (s *Shift) ClockIn() error {
	switch s.Status {
	case "", ShiftStatusScheduled, ShiftStatusNoShow:
		// ok
	default:
		if s.Status != ShiftStatusClockedOut {
			return fmt.Errorf("%w: clock-in from %s", ErrInvalidTransition, s.Status)
		}
	}
	now := time.Now().UTC()
	s.Status = ShiftStatusClockedIn
	s.ClockInAt = now
	s.ActualStart = &now
	s.UpdatedAt = now
	return nil
}

// ClockOut completes an active shift.
func (s *Shift) ClockOut() error {
	if s.Status != ShiftStatusClockedIn && s.Status != ShiftStatusOnBreak && s.Status != ShiftStatusActive {
		return fmt.Errorf("%w: clock-out from %s", ErrInvalidTransition, s.Status)
	}
	now := time.Now().UTC()
	if s.Status == ShiftStatusOnBreak && len(s.Breaks) > 0 {
		last := &s.Breaks[len(s.Breaks)-1]
		if last.EndedAt == nil {
			last.EndedAt = &now
		}
	}
	s.Status = ShiftStatusClockedOut
	s.ClockOutAt = &now
	s.ActualEnd = &now
	s.UpdatedAt = now
	return nil
}

// StartBreak puts the shift on break.
func (s *Shift) StartBreak(reason string) error {
	if s.Status != ShiftStatusClockedIn && s.Status != ShiftStatusActive {
		return fmt.Errorf("%w: break from %s", ErrInvalidTransition, s.Status)
	}
	now := time.Now().UTC()
	s.Status = ShiftStatusOnBreak
	s.Breaks = append(s.Breaks, BreakInterval{StartedAt: now, Reason: reason})
	s.UpdatedAt = now
	return nil
}

// EndBreak resumes work after a break.
func (s *Shift) EndBreak() error {
	if s.Status != ShiftStatusOnBreak || len(s.Breaks) == 0 {
		return fmt.Errorf("%w: end break from %s", ErrInvalidTransition, s.Status)
	}
	now := time.Now().UTC()
	last := &s.Breaks[len(s.Breaks)-1]
	if last.EndedAt != nil {
		return fmt.Errorf("%w: break already ended", ErrConflict)
	}
	last.EndedAt = &now
	s.Status = ShiftStatusClockedIn
	s.UpdatedAt = now
	return nil
}

// Cancel cancels a scheduled shift.
func (s *Shift) Cancel() error {
	if s.Status != ShiftStatusScheduled {
		return fmt.Errorf("%w: cancel from %s", ErrInvalidTransition, s.Status)
	}
	s.Status = ShiftStatusCancelled
	s.UpdatedAt = time.Now().UTC()
	return nil
}
