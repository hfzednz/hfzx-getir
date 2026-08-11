-- Brands / manufacturers for catalog products.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE brand_status AS ENUM (
    'active',
    'inactive',
    'archived'
);

CREATE TABLE brands (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    logo_url        TEXT NOT NULL DEFAULT '',
    website_url     TEXT NOT NULL DEFAULT '',
    country_code    CHAR(2) NOT NULL DEFAULT '',
    external_ref    TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          brand_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_brands_tenant_slug UNIQUE (tenant_id, slug),
    CONSTRAINT chk_brands_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$')
);

COMMENT ON TABLE brands IS 'Brand / manufacturer master data; no pricing or inventory.';
COMMENT ON COLUMN brands.external_ref IS 'Optional external ERP/PIM brand identifier.';
COMMENT ON COLUMN brands.status IS 'Lifecycle: active | inactive | archived.';
