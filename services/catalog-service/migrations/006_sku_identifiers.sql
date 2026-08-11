-- External / internal identifiers for variants (barcodes, GTIN, warehouse codes…).
CREATE TYPE sku_identifier_type AS ENUM (
    'barcode',
    'qr',
    'ean',
    'upc',
    'gtin',
    'internal',
    'supplier',
    'warehouse'
);

CREATE TABLE sku_identifiers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id      UUID NOT NULL REFERENCES variants (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    type            sku_identifier_type NOT NULL,
    value           TEXT NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_sku_identifiers_tenant_type_value UNIQUE (tenant_id, type, value),
    CONSTRAINT chk_sku_identifiers_value CHECK (length(trim(value)) > 0)
);

COMMENT ON TABLE sku_identifiers IS 'SKU codes / barcodes; unique per tenant+type. No inventory quantities.';
COMMENT ON COLUMN sku_identifiers.type IS 'barcode | qr | ean | upc | gtin | internal | supplier | warehouse.';
COMMENT ON COLUMN sku_identifiers.value IS 'Normalized identifier string; uniqueness is (tenant_id, type, value).';
