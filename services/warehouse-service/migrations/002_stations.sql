-- Work stations inside a warehouse (pick / pack / dispatch / qc).
CREATE TYPE station_type AS ENUM (
    'pick',
    'pack',
    'dispatch',
    'qc'
);

CREATE TYPE station_status AS ENUM (
    'available',
    'occupied',
    'offline',
    'maintenance'
);

CREATE TABLE stations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    code            TEXT NOT NULL,
    type            station_type NOT NULL,
    status          station_status NOT NULL DEFAULT 'available',
    name            TEXT NOT NULL DEFAULT '',
    zone_code       TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_stations_warehouse_code UNIQUE (warehouse_id, code),
    CONSTRAINT chk_stations_code CHECK (code ~ '^[A-Za-z0-9][A-Za-z0-9_-]*$')
);

COMMENT ON TABLE stations IS 'Physical workstations for pick/pack/dispatch/qc flows.';
COMMENT ON COLUMN stations.type IS 'pick | pack | dispatch | qc.';
COMMENT ON COLUMN stations.status IS 'available | occupied | offline | maintenance.';
COMMENT ON COLUMN stations.zone_code IS 'Opaque zone hint; not an inventory location FK.';
