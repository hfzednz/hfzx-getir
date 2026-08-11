-- finance-ledger-service: tax_rules, invoices, invoice_lines, credit_notes
CREATE TABLE IF NOT EXISTS tax_rules (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    code         TEXT NOT NULL,
    name         TEXT NOT NULL,
    rate_bps     BIGINT NOT NULL CHECK (rate_bps >= 0 AND rate_bps <= 10000),
    currency     CHAR(3) NOT NULL DEFAULT '',
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS invoices (
    id                UUID PRIMARY KEY,
    tenant_id         UUID NOT NULL,
    status            TEXT NOT NULL CHECK (status IN ('draft','issued','paid','void','credited')),
    currency          CHAR(3) NOT NULL,
    counterparty_ref  TEXT NOT NULL DEFAULT '',
    external_ref      TEXT NOT NULL DEFAULT '',
    idempotency_key   TEXT,
    subtotal_minor    BIGINT NOT NULL DEFAULT 0 CHECK (subtotal_minor >= 0),
    tax_minor         BIGINT NOT NULL DEFAULT 0 CHECK (tax_minor >= 0),
    total_minor       BIGINT NOT NULL DEFAULT 0 CHECK (total_minor >= 0),
    issued_at         TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version           BIGINT NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invoices_tenant_idem
    ON invoices (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

CREATE TABLE IF NOT EXISTS invoice_lines (
    id              UUID PRIMARY KEY,
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    tenant_id       UUID NOT NULL,
    description     TEXT NOT NULL,
    qty             BIGINT NOT NULL CHECK (qty > 0),
    unit_minor      BIGINT NOT NULL CHECK (unit_minor >= 0),
    tax_minor       BIGINT NOT NULL DEFAULT 0 CHECK (tax_minor >= 0),
    total_minor     BIGINT NOT NULL CHECK (total_minor >= 0),
    tax_code        TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS credit_notes (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    invoice_id       UUID NOT NULL REFERENCES invoices(id),
    currency         CHAR(3) NOT NULL,
    amount_minor     BIGINT NOT NULL CHECK (amount_minor > 0),
    reason           TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_credit_notes_tenant_idem
    ON credit_notes (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
