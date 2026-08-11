CREATE TABLE IF NOT EXISTS synonym_rules (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    locale     TEXT NOT NULL DEFAULT '',
    term       TEXT NOT NULL,
    synonyms   TEXT[] NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS boost_rules (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    kind        TEXT NOT NULL,
    product_ids UUID[] NOT NULL DEFAULT '{}',
    category_id UUID,
    brand_id    UUID,
    weight      DOUBLE PRECISION NOT NULL DEFAULT 1,
    priority    INT NOT NULL DEFAULT 0,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS index_jobs (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    mode       TEXT NOT NULL,
    status     TEXT NOT NULL,
    docs_total INT NOT NULL DEFAULT 0,
    docs_done  INT NOT NULL DEFAULT 0,
    error      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS trend_entries (
    tenant_id  UUID NOT NULL,
    kind       TEXT NOT NULL,
    key        TEXT NOT NULL,
    entity_id  UUID,
    score      DOUBLE PRECISION NOT NULL DEFAULT 0,
    region     TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, kind, key)
);

CREATE TABLE IF NOT EXISTS suggest_candidates (
    tenant_id   UUID NOT NULL,
    text        TEXT NOT NULL,
    product_id  UUID,
    category_id UUID,
    weight      DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, text)
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

CREATE TABLE IF NOT EXISTS search_analytics (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    query_id    UUID NOT NULL,
    query_text  TEXT NOT NULL,
    total_hits  INT NOT NULL,
    zero_result BOOLEAN NOT NULL,
    took_ms     BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);
