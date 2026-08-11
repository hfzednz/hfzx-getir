-- Compliance flags: age, hazard, pharmacy, food, country restrictions, certificates.
CREATE TABLE product_compliance (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    age_restriction     INT NOT NULL DEFAULT 0 CHECK (age_restriction >= 0),
    is_hazardous        BOOLEAN NOT NULL DEFAULT FALSE,
    hazard_class        TEXT NOT NULL DEFAULT '',
    is_pharmacy         BOOLEAN NOT NULL DEFAULT FALSE,
    requires_prescription BOOLEAN NOT NULL DEFAULT FALSE,
    is_food             BOOLEAN NOT NULL DEFAULT FALSE,
    is_organic          BOOLEAN NOT NULL DEFAULT FALSE,
    is_halal            BOOLEAN NOT NULL DEFAULT FALSE,
    is_vegan            BOOLEAN NOT NULL DEFAULT FALSE,
    is_gluten_free      BOOLEAN NOT NULL DEFAULT FALSE,
    -- ISO-3166 alpha-2 codes where sale is restricted/blocked.
    restricted_countries TEXT[] NOT NULL DEFAULT '{}',
    allowed_countries   TEXT[] NOT NULL DEFAULT '{}',
    certificates        JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_compliance_product UNIQUE (product_id)
);

COMMENT ON TABLE product_compliance IS 'Regulatory / compliance flags for catalog products.';
COMMENT ON COLUMN product_compliance.age_restriction IS 'Minimum age (0 = none).';
COMMENT ON COLUMN product_compliance.certificates IS 'JSON array of certificate objects (type, issuer, expires_at, url…).';
COMMENT ON COLUMN product_compliance.restricted_countries IS 'Countries where product must not be sold.';
COMMENT ON COLUMN product_compliance.allowed_countries IS 'If non-empty, only these countries are allowed.';
