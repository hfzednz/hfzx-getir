-- Workforce: employees, shifts, attendance, breaks (WH-scoped, not HR payroll).
CREATE TYPE employee_status AS ENUM (
    'active',
    'inactive',
    'suspended'
);

CREATE TYPE employee_role AS ENUM (
    'picker',
    'packer',
    'dispatcher',
    'qc',
    'supervisor',
    'runner',
    'maintenance'
);

CREATE TYPE shift_status AS ENUM (
    'scheduled',
    'active',
    'clocked_in',
    'on_break',
    'clocked_out',
    'completed',
    'cancelled',
    'no_show'
);

CREATE TYPE attendance_event_type AS ENUM (
    'clock_in',
    'clock_out',
    'break_start',
    'break_end'
);

CREATE TABLE employees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    -- Opaque identity principal when linked.
    principal_id    UUID,
    badge_code      TEXT NOT NULL DEFAULT '',
    display_name    TEXT NOT NULL,
    role            employee_role NOT NULL DEFAULT 'picker',
    status          employee_status NOT NULL DEFAULT 'active',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_employees_warehouse_badge
    ON employees (warehouse_id, badge_code)
    WHERE badge_code <> '' AND deleted_at IS NULL;

CREATE TABLE shifts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    employee_id     UUID NOT NULL REFERENCES employees (id) ON DELETE RESTRICT,
    status          shift_status NOT NULL DEFAULT 'scheduled',
    role            employee_role NOT NULL DEFAULT 'picker',
    planned_start   TIMESTAMPTZ,
    planned_end     TIMESTAMPTZ,
    actual_start    TIMESTAMPTZ,
    actual_end      TIMESTAMPTZ,
    clock_in_at     TIMESTAMPTZ,
    clock_out_at    TIMESTAMPTZ,
    station_id      UUID REFERENCES stations (id) ON DELETE SET NULL,
    breaks          JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_shifts_planned_range CHECK (
        planned_start IS NULL OR planned_end IS NULL OR planned_end > planned_start
    )
);

CREATE TABLE attendance_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
    shift_id        UUID REFERENCES shifts (id) ON DELETE SET NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    event_type      attendance_event_type NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    source          TEXT NOT NULL DEFAULT 'mobile',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE break_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_id        UUID NOT NULL REFERENCES shifts (id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at        TIMESTAMPTZ,
    break_kind      TEXT NOT NULL DEFAULT 'rest',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_break_periods_range CHECK (ended_at IS NULL OR ended_at >= started_at)
);

COMMENT ON TABLE employees IS 'WH-scoped workforce roster; not HR/payroll SoT.';
COMMENT ON TABLE shifts IS 'Scheduled/active shifts for station staffing.';
COMMENT ON TABLE attendance_events IS 'Clock in/out and break markers.';
COMMENT ON TABLE break_periods IS 'Open/closed break intervals within a shift.';
