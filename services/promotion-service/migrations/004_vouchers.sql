-- vouchers + redemptions
CREATE TABLE IF NOT EXISTS vouchers (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    promotion_id     UUID REFERENCES promotions(id),
    code             TEXT NOT NULL,
    principal_id     UUID NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('issued','redeemed','expired','void')),
    value_minor      BIGINT NOT NULL CHECK (value_minor >= 0),
    currency         CHAR(3) NOT NULL,
    remaining_minor  BIGINT NOT NULL CHECK (remaining_minor >= 0),
    starts_at        TIMESTAMPTZ,
    ends_at          TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS voucher_redemptions (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    voucher_id       UUID NOT NULL REFERENCES vouchers(id),
    principal_id     UUID NOT NULL,
    idempotency_key  TEXT NOT NULL,
    order_ref        TEXT NOT NULL DEFAULT '',
    amount_minor     BIGINT NOT NULL CHECK (amount_minor >= 0),
    currency         CHAR(3) NOT NULL,
    redeemed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, idempotency_key)
);
