-- Carts: guest or principal shopping cart aggregate.
-- No product master, stock ledger, orders, or PSP ownership.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE cart_status AS ENUM (
    'active',
    'abandoned',
    'converted',
    'merged'
);

CREATE TABLE carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    guest_token     TEXT NOT NULL DEFAULT '',
    principal_id    UUID,
    city_id         UUID,
    status          cart_status NOT NULL DEFAULT 'active',
    currency        CHAR(3) NOT NULL DEFAULT 'TRY',
    version         BIGINT NOT NULL DEFAULT 1,
    merged_into_id  UUID REFERENCES carts (id) ON DELETE SET NULL,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    abandoned_at    TIMESTAMPTZ,
    converted_at    TIMESTAMPTZ,
    merged_at       TIMESTAMPTZ,

    CONSTRAINT chk_carts_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_carts_owner CHECK (
        (guest_token <> '' AND principal_id IS NULL)
        OR (guest_token = '' AND principal_id IS NOT NULL)
        OR status IN ('merged', 'converted', 'abandoned')
    )
);

COMMENT ON TABLE carts IS 'Cart aggregate; opaque variant refs only; prices via PricingClient.';
COMMENT ON COLUMN carts.guest_token IS 'Anonymous shopper token when principal_id is null.';
COMMENT ON COLUMN carts.principal_id IS 'Authenticated principal; opaque identity-service id.';
