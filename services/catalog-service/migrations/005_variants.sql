-- Product variants (option axes: color, size, flavor…).
CREATE TYPE variant_status AS ENUM (
    'draft',
    'active',
    'inactive',
    'archived',
    'deleted'
);

CREATE TABLE variants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    name            TEXT NOT NULL DEFAULT '',
    -- Axis values, e.g. {"color":"red","size":"M"}
    option_values   JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          variant_status NOT NULL DEFAULT 'draft',
    sort_order      INT NOT NULL DEFAULT 0,
    barcode_hint    TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_variants_product_options UNIQUE (product_id, option_values)
);

COMMENT ON TABLE variants IS 'Sellable/variant axes under a product; stock qty owned by inventory-service.';
COMMENT ON COLUMN variants.option_values IS 'JSON map of axis code → value; uniqueness scoped to product.';
COMMENT ON COLUMN variants.status IS 'draft | active | inactive | archived | deleted.';
