-- Equipment registry + heartbeats (scanners, printers, forklifts, robots, iot).
CREATE TYPE equipment_kind AS ENUM (
    'scanner',
    'printer',
    'forklift',
    'robot',
    'iot',
    'scale',
    'conveyor',
    'other'
);

CREATE TYPE equipment_status AS ENUM (
    'online',
    'offline',
    'degraded',
    'maintenance',
    'retired'
);

CREATE TABLE equipment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    station_id      UUID REFERENCES stations (id) ON DELETE SET NULL,
    code            TEXT NOT NULL,
    kind            equipment_kind NOT NULL,
    status          equipment_status NOT NULL DEFAULT 'offline',
    name            TEXT NOT NULL DEFAULT '',
    serial_number   TEXT NOT NULL DEFAULT '',
    firmware        TEXT NOT NULL DEFAULT '',
    last_heartbeat  TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_equipment_warehouse_code UNIQUE (warehouse_id, code),
    CONSTRAINT chk_equipment_code CHECK (code ~ '^[A-Za-z0-9][A-Za-z0-9_-]*$')
);

CREATE TABLE equipment_heartbeats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment (id) ON DELETE CASCADE,
    status          equipment_status NOT NULL,
    battery_pct     INT CHECK (battery_pct IS NULL OR (battery_pct >= 0 AND battery_pct <= 100)),
    signal_rssi     INT,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE equipment IS 'Warehouse device registry; IoT platform core remains external.';
COMMENT ON COLUMN equipment.kind IS 'scanner | printer | forklift | robot | iot | scale | conveyor | other.';
COMMENT ON TABLE equipment_heartbeats IS 'Append-friendly heartbeat samples from edge devices.';
