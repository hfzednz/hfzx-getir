-- Localized product content (titles, descriptions, food/pharma copy).
CREATE TABLE product_locales (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    lang            TEXT NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    subtitle        TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    short_description TEXT NOT NULL DEFAULT '',
    specs           TEXT NOT NULL DEFAULT '',
    usage_text      TEXT NOT NULL DEFAULT '',
    warnings        TEXT NOT NULL DEFAULT '',
    ingredients     TEXT NOT NULL DEFAULT '',
    allergens       TEXT NOT NULL DEFAULT '',
    nutrition       TEXT NOT NULL DEFAULT '',
    storage         TEXT NOT NULL DEFAULT '',
    origin          TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_locales_product_lang UNIQUE (product_id, lang),
    CONSTRAINT chk_product_locales_lang CHECK (lang ~ '^[a-z]{2}(-[A-Z]{2})?$')
);

COMMENT ON TABLE product_locales IS 'i18n product content; no pricing or inventory fields.';
COMMENT ON COLUMN product_locales.lang IS 'BCP-47 language tag, e.g. en, tr, en-US.';
COMMENT ON COLUMN product_locales.usage_text IS 'How-to-use / directions copy.';
