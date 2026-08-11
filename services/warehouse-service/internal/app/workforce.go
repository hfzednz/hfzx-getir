package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// RegisterEmployeeCmd registers warehouse staff.
type RegisterEmployeeCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	ExternalRef string
	DisplayName string
	Role        domain.EmployeeRole
}

// RegisterEmployee creates a workforce member.
func (d *Deps) RegisterEmployee(ctx context.Context, in RegisterEmployeeCmd) (domain.Employee, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.DisplayName == "" {
		return domain.Employee{}, domain.ErrInvalidArgument
	}
	role := in.Role
	if role == "" {
		role = domain.EmployeeRolePicker
	}
	now := d.now()
	e := domain.Employee{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		ExternalRef: in.ExternalRef, DisplayName: in.DisplayName, Role: role,
		Active: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Workforce.CreateEmployee(ctx, e); err != nil {
		return domain.Employee{}, err
	}
	return e, nil
}

// ClockInCmd starts a shift.
type ClockInCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	EmployeeID  uuid.UUID
}

// ClockIn starts attendance for an employee.
func (d *Deps) ClockIn(ctx context.Context, in ClockInCmd) (domain.Shift, error) {
	if _, err := d.Workforce.GetEmployee(ctx, in.TenantID, in.EmployeeID); err != nil {
		return domain.Shift{}, err
	}
	if _, err := d.Workforce.GetActiveShift(ctx, in.TenantID, in.EmployeeID); err == nil {
		return domain.Shift{}, domain.ErrConflict
	}
	now := d.now()
	s := domain.Shift{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		EmployeeID: in.EmployeeID, Status: domain.ShiftStatusClockedIn,
		ClockInAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Workforce.CreateShift(ctx, s); err != nil {
		return domain.Shift{}, err
	}
	return s, nil
}

// ClockOutCmd ends a shift.
type ClockOutCmd struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
}

// ClockOut ends the active shift.
func (d *Deps) ClockOut(ctx context.Context, in ClockOutCmd) (domain.Shift, error) {
	s, err := d.Workforce.GetActiveShift(ctx, in.TenantID, in.EmployeeID)
	if err != nil {
		return domain.Shift{}, err
	}
	now := d.now()
	if s.Status == domain.ShiftStatusOnBreak && len(s.Breaks) > 0 {
		last := &s.Breaks[len(s.Breaks)-1]
		if last.EndedAt == nil {
			last.EndedAt = &now
		}
	}
	s.Status = domain.ShiftStatusClockedOut
	s.ClockOutAt = &now
	s.UpdatedAt = now
	if err := d.Workforce.UpdateShift(ctx, s); err != nil {
		return domain.Shift{}, err
	}
	return s, nil
}

// StartBreakCmd starts a break on the active shift.
type StartBreakCmd struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	Reason     string
}

// StartBreak puts the active shift on break.
func (d *Deps) StartBreak(ctx context.Context, in StartBreakCmd) (domain.Shift, error) {
	s, err := d.Workforce.GetActiveShift(ctx, in.TenantID, in.EmployeeID)
	if err != nil {
		return domain.Shift{}, err
	}
	if s.Status != domain.ShiftStatusClockedIn {
		return domain.Shift{}, domain.ErrInvalidTransition
	}
	now := d.now()
	s.Status = domain.ShiftStatusOnBreak
	s.Breaks = append(s.Breaks, domain.BreakInterval{StartedAt: now, Reason: in.Reason})
	s.UpdatedAt = now
	if err := d.Workforce.UpdateShift(ctx, s); err != nil {
		return domain.Shift{}, err
	}
	return s, nil
}

// EndBreakCmd ends the current break.
type EndBreakCmd struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
}

// EndBreak resumes work after a break.
func (d *Deps) EndBreak(ctx context.Context, in EndBreakCmd) (domain.Shift, error) {
	s, err := d.Workforce.GetActiveShift(ctx, in.TenantID, in.EmployeeID)
	if err != nil {
		return domain.Shift{}, err
	}
	if s.Status != domain.ShiftStatusOnBreak || len(s.Breaks) == 0 {
		return domain.Shift{}, domain.ErrInvalidTransition
	}
	now := d.now()
	last := &s.Breaks[len(s.Breaks)-1]
	if last.EndedAt != nil {
		return domain.Shift{}, domain.ErrInvalidTransition
	}
	last.EndedAt = &now
	s.Status = domain.ShiftStatusClockedIn
	s.UpdatedAt = now
	if err := d.Workforce.UpdateShift(ctx, s); err != nil {
		return domain.Shift{}, err
	}
	return s, nil
}

// PerformanceSnapshot is a simple workforce performance view.
type PerformanceSnapshot struct {
	EmployeeID     uuid.UUID  `json:"employeeId"`
	ActiveShiftID  *uuid.UUID `json:"activeShiftId,omitempty"`
	ShiftStatus    string     `json:"shiftStatus,omitempty"`
	TasksCompleted int        `json:"tasksCompleted"`
}

// WorkforcePerformance returns a simple performance snapshot.
func (d *Deps) WorkforcePerformance(ctx context.Context, tenantID, warehouseID, employeeID uuid.UUID) (PerformanceSnapshot, error) {
	if _, err := d.Workforce.GetEmployee(ctx, tenantID, employeeID); err != nil {
		return PerformanceSnapshot{}, err
	}
	snap := PerformanceSnapshot{EmployeeID: employeeID}
	if s, err := d.Workforce.GetActiveShift(ctx, tenantID, employeeID); err == nil {
		snap.ActiveShiftID = &s.ID
		snap.ShiftStatus = string(s.Status)
	}
	status := domain.TaskStatusCompleted
	list, _, _ := d.Tasks.List(ctx, ports.TaskFilter{
		TenantID: tenantID, WarehouseID: warehouseID, Status: &status, AssigneeID: &employeeID, Limit: 1000,
	})
	snap.TasksCompleted = len(list)
	return snap, nil
}
