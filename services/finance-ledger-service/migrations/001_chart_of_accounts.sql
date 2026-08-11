-- finance-ledger-service: chart of accounts
CREATE TABLE IF NOT EXISTS chart_of_accounts (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    code            TEXT NOT NULL,
    name            TEXT NOT NULL,
    account_type    TEXT NOT NULL CHECK (account_type IN ('asset','liability','revenue','expense','clearing')),
    currency        CHAR(3) NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version         BIGINT NOT NULL DEFAULT 1,
    UNIQUE (tenant_id, code)
);
