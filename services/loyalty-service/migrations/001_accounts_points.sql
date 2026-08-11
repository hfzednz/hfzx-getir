-- Loyalty accounts + point ledger. Points/XP live here; money cashback via wallet.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE loyalty_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    principal_id    UUID NOT NULL,
    points          BIGINT NOT NULL DEFAULT 0,
    tier_points     BIGINT NOT NULL DEFAULT 0,
    xp              BIGINT NOT NULL DEFAULT 0,
    level           INT NOT NULL DEFAULT 1,
    active          BOOLEAN NOT NULL DEFAULT true,
    version         BIGINT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_loyalty_accounts_tenant_principal UNIQUE (tenant_id, principal_id),
    CONSTRAINT chk_loyalty_accounts_points CHECK (points >= 0 AND tier_points >= 0 AND xp >= 0),
    CONSTRAINT chk_loyalty_accounts_version CHECK (version >= 1)
);

CREATE TYPE loyalty_point_kind AS ENUM (
    'earn',
    'redeem',
    'expire',
    'adjust',
    'grant'
);

CREATE TABLE point_ledger (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    account_id        UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    kind              loyalty_point_kind NOT NULL,
    points            BIGINT NOT NULL,
    balance_after     BIGINT NOT NULL,
    order_id          UUID,
    reference         TEXT NOT NULL DEFAULT '',
    idempotency_key   TEXT NOT NULL,
    metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_point_ledger_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_point_ledger_points CHECK (points >= 0),
    CONSTRAINT chk_point_ledger_balance CHECK (balance_after >= 0)
);

CREATE TABLE account_stats (
    tenant_id   UUID NOT NULL,
    account_id  UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    stat_key    TEXT NOT NULL,
    value       BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, account_id, stat_key),
    CONSTRAINT chk_account_stats_value CHECK (value >= 0)
);
