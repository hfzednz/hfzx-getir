CREATE TABLE IF NOT EXISTS quality_scores (
    id         UUID PRIMARY KEY,
    review_id  UUID NOT NULL REFERENCES reviews(id),
    tenant_id  UUID NOT NULL,
    dimension  TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (review_id, dimension)
);

CREATE TABLE IF NOT EXISTS moderation_cases (
    id             UUID PRIMARY KEY,
    review_id      UUID NOT NULL REFERENCES reviews(id),
    tenant_id      UUID NOT NULL,
    status         TEXT NOT NULL,
    auto_decision  TEXT NOT NULL DEFAULT '',
    ai_score       DOUBLE PRECISION NOT NULL DEFAULT 0,
    labels         TEXT[] NOT NULL DEFAULT '{}',
    fraud_score    DOUBLE PRECISION NOT NULL DEFAULT 0,
    fraud_signals  TEXT[] NOT NULL DEFAULT '{}',
    pii_masked     BOOLEAN NOT NULL DEFAULT FALSE,
    assignee_id    UUID,
    decision_note  TEXT NOT NULL DEFAULT '',
    decided_by     UUID,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    decided_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_moderation_pending
    ON moderation_cases (tenant_id, status, created_at);
