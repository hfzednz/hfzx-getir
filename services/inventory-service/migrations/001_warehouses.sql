-- Warehouses: physical or logical stock sites owned by a tenant.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE warehouse_status AS ENUM (
    'active',
    'inactive',
    'maintenance',
    'closed'
);

CREATE TABLE warehouses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    region_id       UUID,
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    status          warehouse_status NOT NULL DEFAULT 'active',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_warehouses_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_warehouses_code CHECK (code ~ '^[A-Za-z0-9][A-Za-z0-9_-]*$'),
    CONSTRAINT chk_warehouses_timezone CHECK (timezone <> '')
);

COMMENT ON TABLE warehouses IS 'Inventory warehouse / DC sites; no catalog or pricing data.';
COMMENT ON COLUMN warehouses.region_id IS 'Opaque region identifier for ATP aggregation; not owned here.';
COMMENT ON COLUMN warehouses.timezone IS 'IANA timezone used for expiry/cutoff calculations.';
COMMENT ON COLUMN warehouses.status IS 'Lifecycle: active | inactive | maintenance | closed.';
