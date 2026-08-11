-- Attribute values attached to products (schema-driven via attribute_defs).
CREATE TABLE product_attributes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    attribute_def_id    UUID NOT NULL REFERENCES attribute_defs (id) ON DELETE RESTRICT,
    tenant_id           UUID NOT NULL,
    -- Typed payload: {"value": ...} or structured per attribute type.
    value               JSONB NOT NULL DEFAULT '{}'::jsonb,
    locale              TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_attributes_def_locale UNIQUE (product_id, attribute_def_id, locale)
);

COMMENT ON TABLE product_attributes IS 'Product attribute values; schema validated against attribute_defs.';
COMMENT ON COLUMN product_attributes.value IS 'JSON value conforming to attribute_defs.schema / type.';
COMMENT ON COLUMN product_attributes.locale IS 'Empty = locale-agnostic; otherwise BCP-47 language tag.';
