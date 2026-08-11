-- reviews aggregate
CREATE TABLE IF NOT EXISTS reviews (
    id                UUID PRIMARY KEY,
    tenant_id         UUID NOT NULL,
    author_id         UUID NOT NULL,
    target_type       TEXT NOT NULL,
    target_id         UUID NOT NULL,
    order_id          UUID,
    locale            TEXT NOT NULL DEFAULT 'tr-TR',
    title             TEXT NOT NULL DEFAULT '',
    body              TEXT NOT NULL DEFAULT '',
    anonymous         BOOLEAN NOT NULL DEFAULT FALSE,
    verified_purchase BOOLEAN NOT NULL DEFAULT FALSE,
    verified_delivery BOOLEAN NOT NULL DEFAULT FALSE,
    status            TEXT NOT NULL,
    sentiment         DOUBLE PRECISION NOT NULL DEFAULT 0,
    topics            TEXT[] NOT NULL DEFAULT '{}',
    tags              TEXT[] NOT NULL DEFAULT '{}',
    helpful_count     INT NOT NULL DEFAULT 0,
    not_helpful_count INT NOT NULL DEFAULT 0,
    report_count      INT NOT NULL DEFAULT 0,
    pinned            BOOLEAN NOT NULL DEFAULT FALSE,
    revision          INT NOT NULL DEFAULT 1,
    idempotency_key   TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    published_at      TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_reviews_idempotency
    ON reviews (tenant_id, idempotency_key) WHERE idempotency_key <> '';
CREATE INDEX IF NOT EXISTS idx_reviews_target
    ON reviews (tenant_id, target_type, target_id, status);
CREATE INDEX IF NOT EXISTS idx_reviews_author
    ON reviews (tenant_id, author_id, created_at DESC);
