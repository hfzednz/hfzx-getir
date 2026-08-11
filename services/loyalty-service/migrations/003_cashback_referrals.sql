-- Cashback grants (wallet credit orchestration) + referrals.

CREATE TYPE cashback_status AS ENUM (
    'pending',
    'issued',
    'failed'
);

CREATE TABLE cashback_grants (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    account_id        UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    principal_id      UUID NOT NULL,
    amount_minor      BIGINT NOT NULL,
    currency          CHAR(3) NOT NULL DEFAULT 'TRY',
    account_type      TEXT NOT NULL DEFAULT 'cashback',
    status            cashback_status NOT NULL DEFAULT 'pending',
    order_id          UUID,
    idempotency_key   TEXT NOT NULL,
    wallet_ref        TEXT NOT NULL DEFAULT '',
    failure_reason    TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_cashback_grants_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_cashback_amount CHECK (amount_minor > 0),
    CONSTRAINT chk_cashback_currency CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE TABLE referral_codes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    account_id    UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    principal_id  UUID NOT NULL,
    code          TEXT NOT NULL,
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_referral_codes_code UNIQUE (tenant_id, code),
    CONSTRAINT uq_referral_codes_account UNIQUE (tenant_id, account_id)
);

CREATE TYPE referral_status AS ENUM (
    'open',
    'applied',
    'completed',
    'rejected'
);

CREATE TABLE referral_events (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID NOT NULL,
    code_id            UUID NOT NULL REFERENCES referral_codes (id),
    referrer_account   UUID NOT NULL REFERENCES loyalty_accounts (id),
    referee_account    UUID NOT NULL REFERENCES loyalty_accounts (id),
    referee_principal  UUID NOT NULL,
    status             referral_status NOT NULL DEFAULT 'applied',
    order_id           UUID,
    points_granted     BIGINT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_referral_events_referee UNIQUE (tenant_id, referee_account),
    CONSTRAINT chk_referral_not_self CHECK (referrer_account <> referee_account),
    CONSTRAINT chk_referral_points CHECK (points_granted >= 0)
);
