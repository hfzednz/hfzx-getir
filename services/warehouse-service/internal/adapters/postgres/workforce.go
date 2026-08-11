package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

const metaExternalRef = "_externalRef"

// WorkforceRepo persists employees and shifts.
type WorkforceRepo struct{ DB *sql.DB }

var _ ports.WorkforceRepo = (*WorkforceRepo)(nil)

func (r *WorkforceRepo) CreateEmployee(ctx context.Context, e domain.Employee) error {
	if err := ensureWarehouse(ctx, r.DB, e.TenantID, e.WarehouseID); err != nil {
		return err
	}
	status := e.Status
	if status == "" {
		if e.Active {
			status = domain.EmployeeStatusActive
		} else {
			status = domain.EmployeeStatusInactive
		}
	}
	role := e.Role
	if role == "" {
		role = domain.EmployeeRolePicker
	}
	meta := mergeMeta(e.Metadata, map[string]any{})
	if e.ExternalRef != "" {
		meta[metaExternalRef] = e.ExternalRef
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO employees (
			id, tenant_id, warehouse_id, principal_id, badge_code, display_name, role, status,
			metadata, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		e.ID, e.TenantID, e.WarehouseID, nullUUID(e.PrincipalID), e.BadgeCode, e.DisplayName,
		string(role), string(status), meta, e.CreatedAt, e.UpdatedAt, nullTime(e.DeletedAt),
	)
	return mapUniqueViolation(err)
}

func (r *WorkforceRepo) GetEmployee(ctx context.Context, tenantID, id uuid.UUID) (domain.Employee, error) {
	return r.scanEmployee(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, principal_id, badge_code, display_name, role, status,
			metadata, created_at, updated_at, deleted_at
		FROM employees WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *WorkforceRepo) ListEmployees(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Employee, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, warehouse_id, principal_id, badge_code, display_name, role, status,
			metadata, created_at, updated_at, deleted_at
		FROM employees
		WHERE tenant_id=$1 AND warehouse_id=$2 AND deleted_at IS NULL
		ORDER BY display_name`, tenantID, warehouseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Employee, 0)
	for rows.Next() {
		e, err := scanEmployeeRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *WorkforceRepo) CreateShift(ctx context.Context, s domain.Shift) error {
	if err := ensureWarehouse(ctx, r.DB, s.TenantID, s.WarehouseID); err != nil {
		return err
	}
	status := s.Status
	if status == "" {
		status = domain.ShiftStatusScheduled
	}
	role := s.Role
	if role == "" {
		role = domain.EmployeeRolePicker
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO shifts (
			id, tenant_id, warehouse_id, employee_id, status, role,
			planned_start, planned_end, actual_start, actual_end, clock_in_at, clock_out_at,
			station_id, breaks, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,
			$13,$14,$15,$16,$17
		)`,
		s.ID, s.TenantID, s.WarehouseID, s.EmployeeID, string(status), string(role),
		nullTimeValue(s.PlannedStart), nullTimeValue(s.PlannedEnd), nullTime(s.ActualStart), nullTime(s.ActualEnd),
		nullTimeValue(s.ClockInAt), nullTime(s.ClockOutAt),
		nullUUID(s.StationID), JSONRaw{V: s.Breaks}, JSONMap(metaGetMap(s.Metadata)), s.CreatedAt, s.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *WorkforceRepo) UpdateShift(ctx context.Context, s domain.Shift) error {
	status := s.Status
	if status == "" {
		status = domain.ShiftStatusScheduled
	}
	role := s.Role
	if role == "" {
		role = domain.EmployeeRolePicker
	}
	res, err := r.DB.ExecContext(ctx, `
		UPDATE shifts SET
			status=$1, role=$2, planned_start=$3, planned_end=$4, actual_start=$5, actual_end=$6,
			clock_in_at=$7, clock_out_at=$8, station_id=$9, breaks=$10, metadata=$11, updated_at=$12
		WHERE id=$13 AND tenant_id=$14`,
		string(status), string(role), nullTimeValue(s.PlannedStart), nullTimeValue(s.PlannedEnd),
		nullTime(s.ActualStart), nullTime(s.ActualEnd), nullTimeValue(s.ClockInAt), nullTime(s.ClockOutAt),
		nullUUID(s.StationID), JSONRaw{V: s.Breaks}, JSONMap(metaGetMap(s.Metadata)), s.UpdatedAt,
		s.ID, s.TenantID,
	)
	return rowsAffectedOrNotFound(res, err)
}

func (r *WorkforceRepo) GetActiveShift(ctx context.Context, tenantID, employeeID uuid.UUID) (domain.Shift, error) {
	return r.scanShift(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, employee_id, status, role,
			planned_start, planned_end, actual_start, actual_end, clock_in_at, clock_out_at,
			station_id, breaks, metadata, created_at, updated_at
		FROM shifts
		WHERE tenant_id=$1 AND employee_id=$2
			AND status IN ('active','clocked_in','on_break')
		ORDER BY updated_at DESC
		LIMIT 1`, tenantID, employeeID))
}

func (r *WorkforceRepo) GetShift(ctx context.Context, tenantID, id uuid.UUID) (domain.Shift, error) {
	return r.scanShift(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, warehouse_id, employee_id, status, role,
			planned_start, planned_end, actual_start, actual_end, clock_in_at, clock_out_at,
			station_id, breaks, metadata, created_at, updated_at
		FROM shifts WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *WorkforceRepo) scanEmployee(row scannable) (domain.Employee, error) {
	e, err := scanEmployeeRow(row)
	if err != nil {
		return domain.Employee{}, mapNotFound(err)
	}
	return e, nil
}

func scanEmployeeRow(row scannable) (domain.Employee, error) {
	var e domain.Employee
	var principalID uuid.NullUUID
	var role, status string
	var meta JSONMap
	var deleted sql.NullTime
	err := row.Scan(
		&e.ID, &e.TenantID, &e.WarehouseID, &principalID, &e.BadgeCode, &e.DisplayName, &role, &status,
		&meta, &e.CreatedAt, &e.UpdatedAt, &deleted,
	)
	if err != nil {
		return domain.Employee{}, err
	}
	e.PrincipalID = scanNullUUID(principalID)
	e.Role = domain.EmployeeRole(role)
	e.Status = domain.EmployeeStatus(status)
	e.Active = e.Status == domain.EmployeeStatusActive
	e.DeletedAt = scanNullTime(deleted)
	e.CreatedAt = e.CreatedAt.UTC()
	e.UpdatedAt = e.UpdatedAt.UTC()
	userMeta := map[string]any{}
	for k, v := range meta {
		if k == metaExternalRef {
			if str, ok := v.(string); ok {
				e.ExternalRef = str
			}
			continue
		}
		userMeta[k] = v
	}
	e.Metadata = userMeta
	return e, nil
}

func (r *WorkforceRepo) scanShift(row scannable) (domain.Shift, error) {
	s, err := scanShiftRow(row)
	if err != nil {
		return domain.Shift{}, mapNotFound(err)
	}
	return s, nil
}

func scanShiftRow(row scannable) (domain.Shift, error) {
	var s domain.Shift
	var status, role string
	var plannedStart, plannedEnd, actualStart, actualEnd, clockIn, clockOut sql.NullTime
	var stationID uuid.NullUUID
	var breaksRaw []byte
	var meta JSONMap
	err := row.Scan(
		&s.ID, &s.TenantID, &s.WarehouseID, &s.EmployeeID, &status, &role,
		&plannedStart, &plannedEnd, &actualStart, &actualEnd, &clockIn, &clockOut,
		&stationID, &breaksRaw, &meta, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return domain.Shift{}, err
	}
	s.Status = domain.ShiftStatus(status)
	s.Role = domain.EmployeeRole(role)
	s.PlannedStart = scanTimeOrZero(plannedStart)
	s.PlannedEnd = scanTimeOrZero(plannedEnd)
	s.ActualStart = scanNullTime(actualStart)
	s.ActualEnd = scanNullTime(actualEnd)
	s.ClockInAt = scanTimeOrZero(clockIn)
	s.ClockOutAt = scanNullTime(clockOut)
	s.StationID = scanNullUUID(stationID)
	s.Metadata = map[string]any(meta)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	if len(breaksRaw) > 0 && string(breaksRaw) != "null" {
		_ = json.Unmarshal(breaksRaw, &s.Breaks)
	}
	if s.Breaks == nil {
		s.Breaks = []domain.BreakInterval{}
	}
	return s, nil
}
