-- Price books and scoped price entries (opaque variant_id; no catalog ownership).
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE price_scope AS ENUM (
    'base',
    'regional',
    'warehouse',
    'customer',
    'vip',
    'corporate'
);

CREATE TABLE price_books (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    currency    CHAR(3) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_price_books_currency CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE TABLE price_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    price_book_id   UUID NOT NULL REFERENCES price_books (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL,
    scope           price_scope NOT NULL,
    scope_id        UUID,
    amount_minor    BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    valid_from      TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_to        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_price_entries_amount CHECK (amount_minor >= 0),
    CONSTRAINT chk_price_entries_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_price_entries_validity CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT chk_price_entries_scope_id CHECK (
        (scope = 'base' AND scope_id IS NULL)
        OR (scope <> 'base' AND scope_id IS NOT NULL)
    )
);

COMMENT ON TABLE price_books IS 'Tenant price books; currency scoped.';
COMMENT ON TABLE price_entries IS 'Waterfall prices: base→regional→warehouse→customer→vip→corporate.';
COMMENT ON COLUMN price_entries.variant_id IS 'Opaque catalog variant; no FK to catalog-service.';
