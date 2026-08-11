-- Quote snapshots from PricingClient (all money in integer minor units).
CREATE TABLE cart_quotes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id             UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    quote_id            UUID NOT NULL, -- opaque pricing quote id
    currency            CHAR(3) NOT NULL,
    subtotal_minor      BIGINT NOT NULL DEFAULT 0,
    discount_minor      BIGINT NOT NULL DEFAULT 0,
    tax_minor           BIGINT NOT NULL DEFAULT 0,
    delivery_minor      BIGINT NOT NULL DEFAULT 0,
    service_minor       BIGINT NOT NULL DEFAULT 0,
    packaging_minor     BIGINT NOT NULL DEFAULT 0,
    tip_minor           BIGINT NOT NULL DEFAULT 0,
    total_minor         BIGINT NOT NULL DEFAULT 0,
    line_quotes         JSONB NOT NULL DEFAULT '[]'::jsonb,
    quoted_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_cart_quotes_cart UNIQUE (cart_id),
    CONSTRAINT chk_cart_quotes_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_cart_quotes_money_nonneg CHECK (
        subtotal_minor >= 0
        AND discount_minor >= 0
        AND tax_minor >= 0
        AND delivery_minor >= 0
        AND service_minor >= 0
        AND packaging_minor >= 0
        AND tip_minor >= 0
        AND total_minor >= 0
    )
);

COMMENT ON TABLE cart_quotes IS 'Last pricing quote snapshot; refreshed via PricingClient.Quote.';
