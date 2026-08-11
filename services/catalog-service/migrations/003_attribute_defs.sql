-- Schema-driven attribute definitions (PIM attribute catalog).
CREATE TYPE attribute_type AS ENUM (
    'text',
    'number',
    'boolean',
    'date',
    'color',
    'size',
    'weight',
    'dimension',
    'volume',
    'energy',
    'nutrition',
    'custom'
);

CREATE TABLE attribute_defs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    type            attribute_type NOT NULL DEFAULT 'text',
    -- Validation / UI schema: units, enum options, min/max, locale labels, etc.
    schema          JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_required     BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable   BOOLEAN NOT NULL DEFAULT FALSE,
    is_variant_axis BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_attribute_defs_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_attribute_defs_code CHECK (code ~ '^[a-z][a-z0-9_]{1,63}$')
);

COMMENT ON TABLE attribute_defs IS 'Tenant-scoped attribute definitions; values live on product_attributes.';
COMMENT ON COLUMN attribute_defs.type IS 'text | number | boolean | date | color | size | weight | dimension | volume | energy | nutrition | custom.';
COMMENT ON COLUMN attribute_defs.schema IS 'JSON schema / UI hints for validation and rendering.';
COMMENT ON COLUMN attribute_defs.is_variant_axis IS 'When true, attribute may appear in variant option_values.';
