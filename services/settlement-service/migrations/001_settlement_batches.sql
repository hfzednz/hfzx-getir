-- settlement-service: settlement_batches + settlement_lines
CREATE TABLE IF NOT EXISTS settlement_batches (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    status           TEXT NOT NULL CHECK (status IN (
        'draft','pending_approval','approved','paying','completed','failed'
    )),
    currency         CHAR(3) NOT NULL,
    period_start     TIMESTAMPTZ NOT NULL,
    period_end       TIMESTAMPTZ NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT,
    total_minor      BIGINT NOT NULL DEFAULT 0 CHECK (total_minor >= 0),
    submitted_by     UUID,
    submitted_at     TIMESTAMPTZ,
    approved_by      UUID,
    approved_at      TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    failed_at        TIMESTAMPTZ,
    failure_reason   TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version          BIGINT NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_settlement_batches_tenant_idem
    ON settlement_batches (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';

CREATE TABLE IF NOT EXISTS settlement_lines (
    id              UUID PRIMARY KEY,
    batch_id        UUID NOT NULL REFERENCES settlement_batches(id),
    tenant_id       UUID NOT NULL,
    payee_type      TEXT NOT NULL CHECK (payee_type IN ('courier','supplier','merchant','partner')),
    payee_ref       TEXT NOT NULL,
    amount_minor    BIGINT NOT NULL CHECK (amount_minor > 0),
    currency        CHAR(3) NOT NULL,
    external_ref    TEXT NOT NULL DEFAULT '',
    memo            TEXT NOT NULL DEFAULT ''
);
