CREATE TABLE IF NOT EXISTS product_features (
    product_id  UUID PRIMARY KEY,
    category_id UUID,
    brand_id    UUID,
    tags        TEXT[] NOT NULL DEFAULT '{}',
    price_minor BIGINT NOT NULL DEFAULT 0,
    popularity  DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_avg  DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS behavior_signals (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    user_id    UUID NOT NULL,
    product_id UUID NOT NULL,
    kind       TEXT NOT NULL,
    weight     DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_signals_user ON behavior_signals (tenant_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_product ON behavior_signals (tenant_id, product_id);

CREATE TABLE IF NOT EXISTS co_occurrences (
    tenant_id  UUID NOT NULL,
    product_a  UUID NOT NULL,
    product_b  UUID NOT NULL,
    count      INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, product_a, product_b)
);

CREATE TABLE IF NOT EXISTS outbox_messages (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    aggregate_id UUID NOT NULL,
    topic        TEXT NOT NULL,
    key          TEXT NOT NULL,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL,
    attempts     INT NOT NULL DEFAULT 0,
    last_error   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ
);
