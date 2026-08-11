-- finance-ledger-service: journals + journal_lines (debit/credit minor units)
CREATE TABLE IF NOT EXISTS journals (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('draft','posted')),
    currency         CHAR(3) NOT NULL,
    reference        TEXT NOT NULL DEFAULT '',
    description      TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT,
    posted_at        TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version          BIGINT NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_journals_tenant_idem
    ON journals (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

CREATE TABLE IF NOT EXISTS journal_lines (
    id              UUID PRIMARY KEY,
    journal_id      UUID NOT NULL REFERENCES journals(id),
    tenant_id       UUID NOT NULL,
    account_id      UUID NOT NULL REFERENCES chart_of_accounts(id),
    account_code    TEXT NOT NULL DEFAULT '',
    debit_minor     BIGINT NOT NULL DEFAULT 0 CHECK (debit_minor >= 0),
    credit_minor    BIGINT NOT NULL DEFAULT 0 CHECK (credit_minor >= 0),
    currency        CHAR(3) NOT NULL,
    memo            TEXT NOT NULL DEFAULT '',
    CONSTRAINT journal_lines_xor_dc CHECK (
        (debit_minor > 0 AND credit_minor = 0) OR (credit_minor > 0 AND debit_minor = 0)
    )
);
