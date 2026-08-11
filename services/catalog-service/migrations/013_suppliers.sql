-- Suppliers and product links (metadata only). Optional cost_hint is a supplier quote, NOT sell price.
CREATE TYPE supplier_status AS ENUM (
    'active',
    'inactive',
    'archived'
);

CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    contact_email   TEXT NOT NULL DEFAULT '',
    contact_phone   TEXT NOT NULL DEFAULT '',
    country_code    CHAR(2) NOT NULL DEFAULT '',
    external_ref    TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          supplier_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_suppliers_tenant_code UNIQUE (tenant_id, code)
);

CREATE TABLE supplier_products (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id         UUID NOT NULL REFERENCES suppliers (id) ON DELETE CASCADE,
    product_id          UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id          UUID REFERENCES variants (id) ON DELETE SET NULL,
    tenant_id           UUID NOT NULL,
    supplier_sku        TEXT NOT NULL DEFAULT '',
    -- Optional supplier quote in minor currency units (e.g. kuruş). NOT a sellable price.
    cost_hint_minor     BIGINT,
    cost_hint_currency  CHAR(3) NOT NULL DEFAULT '',
    lead_time_days      INT,
    moq                 INT,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_preferred        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_supplier_products_cost_hint CHECK (cost_hint_minor IS NULL OR cost_hint_minor >= 0)
);

CREATE UNIQUE INDEX uq_supplier_products
    ON supplier_products (
        supplier_id,
        product_id,
        COALESCE(variant_id, '00000000-0000-0000-0000-000000000000'::uuid)
    );

COMMENT ON TABLE suppliers IS 'Supplier master; settlement contracts owned by finance.';
COMMENT ON TABLE supplier_products IS 'Product↔supplier link metadata only.';
COMMENT ON COLUMN supplier_products.cost_hint_minor IS 'Supplier quote in minor units — NOT sell price (pricing-service owns sellable prices).';
COMMENT ON COLUMN supplier_products.cost_hint_currency IS 'ISO-4217 currency for cost_hint_minor.';
