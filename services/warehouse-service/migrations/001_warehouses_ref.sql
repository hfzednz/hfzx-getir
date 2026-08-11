-- Local warehouse config projection. id may match inventory-service warehouse id.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE warehouse_type AS ENUM (
    'dark_store',
    'regional',
    'hub',
    'micro_fulfillment',
    'partner'
);

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
    name            TEXT NOT NULL DEFAULT '',
    type            warehouse_type NOT NULL DEFAULT 'dark_store',
    status          warehouse_status NOT NULL DEFAULT 'active',
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_warehouses_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_warehouses_code CHECK (code ~ '^[A-Za-z0-9][A-Za-z0-9_-]*$'),
    CONSTRAINT chk_warehouses_timezone CHECK (timezone <> '')
);

COMMENT ON TABLE warehouses IS 'Local WH ops config; id may align with inventory warehouse id (opaque, not FK).';
COMMENT ON COLUMN warehouses.type IS 'dark_store | regional | hub | micro_fulfillment | partner.';
COMMENT ON COLUMN warehouses.status IS 'active | inactive | maintenance | closed.';
