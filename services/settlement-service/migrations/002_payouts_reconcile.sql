-- settlement-service: payout_instructions, reconciliations, mismatches
CREATE TABLE IF NOT EXISTS payout_instructions (
    id              UUID PRIMARY KEY,
    batch_id        UUID NOT NULL REFERENCES settlement_batches(id),
    line_id         UUID NOT NULL,
    tenant_id       UUID NOT NULL,
    payee_type      TEXT NOT NULL CHECK (payee_type IN ('courier','supplier','merchant','partner')),
    payee_ref       TEXT NOT NULL,
    amount_minor    BIGINT NOT NULL CHECK (amount_minor > 0),
    currency        CHAR(3) NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('pending','sent','succeeded','failed')),
    provider_ref    TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reconciliations (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    batch_id         UUID NOT NULL REFERENCES settlement_batches(id),
    provider_ref     TEXT NOT NULL DEFAULT '',
    expected_minor   BIGINT NOT NULL,
    reported_minor   BIGINT NOT NULL,
    matched          BOOLEAN NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mismatches (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    batch_id         UUID NOT NULL REFERENCES settlement_batches(id),
    reconcile_id     UUID NOT NULL REFERENCES reconciliations(id),
    expected_minor   BIGINT NOT NULL,
    reported_minor   BIGINT NOT NULL,
    delta_minor      BIGINT NOT NULL,
    detail           TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
