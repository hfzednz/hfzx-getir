-- Wallet core: wallets, accounts (cash|refund|promo|cashback|gift), entries, holds, transfers, limits.
-- Available = balance_minor - held_minor. Money = BIGINT minor units.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE wallet_account_type AS ENUM (
    'cash',
    'refund',
    'promo',
    'cashback',
    'gift'
);

CREATE TYPE wallet_entry_kind AS ENUM (
    'credit',
    'debit',
    'hold',
    'release',
    'adjust'
);

CREATE TYPE wallet_hold_status AS ENUM (
    'active',
    'released',
    'captured'
);

CREATE TABLE wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    principal_id    UUID NOT NULL,
    currency        CHAR(3) NOT NULL DEFAULT 'TRY',
    active          BOOLEAN NOT NULL DEFAULT true,
    version         BIGINT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallets_tenant_principal UNIQUE (tenant_id, principal_id),
    CONSTRAINT chk_wallets_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_wallets_version CHECK (version >= 1)
);

CREATE TABLE wallet_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    account_type    wallet_account_type NOT NULL,
    balance_minor   BIGINT NOT NULL DEFAULT 0,
    held_minor      BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL,
    version         BIGINT NOT NULL DEFAULT 1,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallet_accounts_type UNIQUE (wallet_id, account_type),
    CONSTRAINT chk_wallet_accounts_balance CHECK (balance_minor >= 0 AND held_minor >= 0),
    CONSTRAINT chk_wallet_accounts_available CHECK (balance_minor >= held_minor),
    CONSTRAINT chk_wallet_accounts_currency CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE TABLE wallet_entries (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id         UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    account_id        UUID NOT NULL REFERENCES wallet_accounts (id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL,
    kind              wallet_entry_kind NOT NULL,
    amount_minor      BIGINT NOT NULL,
    currency          CHAR(3) NOT NULL,
    balance_after     BIGINT NOT NULL,
    held_after        BIGINT NOT NULL,
    reference         TEXT NOT NULL DEFAULT '',
    idempotency_key   TEXT NOT NULL,
    metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallet_entries_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_wallet_entries_amount CHECK (amount_minor >= 0)
);

CREATE TABLE wallet_holds (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id         UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    account_id        UUID NOT NULL REFERENCES wallet_accounts (id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL,
    amount_minor      BIGINT NOT NULL,
    currency          CHAR(3) NOT NULL,
    status            wallet_hold_status NOT NULL DEFAULT 'active',
    reference         TEXT NOT NULL DEFAULT '',
    idempotency_key   TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    released_at       TIMESTAMPTZ,

    CONSTRAINT uq_wallet_holds_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_wallet_holds_amount CHECK (amount_minor > 0)
);

CREATE TABLE wallet_transfers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID NOT NULL,
    from_wallet_id     UUID NOT NULL REFERENCES wallets (id),
    from_account_id    UUID NOT NULL REFERENCES wallet_accounts (id),
    to_wallet_id       UUID NOT NULL REFERENCES wallets (id),
    to_account_id      UUID NOT NULL REFERENCES wallet_accounts (id),
    amount_minor       BIGINT NOT NULL,
    currency           CHAR(3) NOT NULL,
    idempotency_key    TEXT NOT NULL,
    reference          TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallet_transfers_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_wallet_transfers_amount CHECK (amount_minor > 0)
);

CREATE TABLE wallet_limits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    limit_type      TEXT NOT NULL,
    amount_minor    BIGINT NOT NULL,
    currency        CHAR(3) NOT NULL,
    window_key      TEXT NOT NULL DEFAULT '',
    used_minor      BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallet_limits UNIQUE (wallet_id, limit_type, window_key),
    CONSTRAINT chk_wallet_limits_amount CHECK (amount_minor >= 0 AND used_minor >= 0)
);
