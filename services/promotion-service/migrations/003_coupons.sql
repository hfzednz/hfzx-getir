-- coupons + redemptions
CREATE TABLE IF NOT EXISTS coupons (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    promotion_id     UUID NOT NULL REFERENCES promotions(id),
    code             TEXT NOT NULL,
    kind             TEXT NOT NULL CHECK (kind IN ('single','multi','personal','public')),
    max_redemptions  INT NOT NULL DEFAULT 0,
    redeemed_count   INT NOT NULL DEFAULT 0,
    principal_id     UUID,
    starts_at        TIMESTAMPTZ,
    ends_at          TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    coupon_id        UUID NOT NULL REFERENCES coupons(id),
    principal_id     UUID NOT NULL,
    idempotency_key  TEXT NOT NULL,
    order_ref        TEXT NOT NULL DEFAULT '',
    discount_minor   BIGINT NOT NULL DEFAULT 0,
    currency         CHAR(3) NOT NULL,
    redeemed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, idempotency_key)
);
