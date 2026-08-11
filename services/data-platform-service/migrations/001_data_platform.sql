CREATE TABLE IF NOT EXISTS event_schemas (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    name          TEXT NOT NULL,
    family        TEXT NOT NULL,
    version       INT NOT NULL,
    compatibility TEXT NOT NULL,
    json_schema   JSONB NOT NULL DEFAULT '{}',
    active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, name, version)
);

CREATE TABLE IF NOT EXISTS analytics_events (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    family          TEXT NOT NULL,
    schema_version  INT NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL DEFAULT '',
    occurred_at     TIMESTAMPTZ NOT NULL,
    ingested_at     TIMESTAMPTZ NOT NULL,
    user_id         UUID,
    session_id      UUID,
    city_id         UUID,
    payload         JSONB NOT NULL DEFAULT '{}',
    payload_hash    TEXT NOT NULL,
    layer           TEXT NOT NULL,
    valid           BOOLEAN NOT NULL,
    error           TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_events_idem
    ON analytics_events (tenant_id, idempotency_key) WHERE idempotency_key <> '';
CREATE INDEX IF NOT EXISTS idx_events_name ON analytics_events (tenant_id, name, ingested_at DESC);

CREATE TABLE IF NOT EXISTS stream_jobs (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    name         TEXT NOT NULL,
    event_name   TEXT NOT NULL,
    window_sec   INT NOT NULL,
    metric_field TEXT NOT NULL DEFAULT '',
    agg          TEXT NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS aggregate_windows (
    tenant_id     UUID NOT NULL,
    job_id        UUID NOT NULL,
    window_start  TIMESTAMPTZ NOT NULL,
    window_end    TIMESTAMPTZ NOT NULL,
    value         DOUBLE PRECISION NOT NULL,
    count         BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, job_id, window_start)
);

CREATE TABLE IF NOT EXISTS lake_datasets (
    id             UUID PRIMARY KEY,
    tenant_id      UUID NOT NULL,
    name           TEXT NOT NULL,
    layer          TEXT NOT NULL,
    format         TEXT NOT NULL,
    location       TEXT NOT NULL,
    partition_by   TEXT[] NOT NULL DEFAULT '{}',
    retention_days INT NOT NULL DEFAULT 90,
    updated_at     TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS fact_snapshots (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    fact_table TEXT NOT NULL,
    grain_key  TEXT NOT NULL,
    measures   JSONB NOT NULL,
    dims       JSONB NOT NULL,
    as_of      TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS kpi_values (
    tenant_id UUID NOT NULL,
    key       TEXT NOT NULL,
    value     DOUBLE PRECISION NOT NULL,
    unit      TEXT NOT NULL,
    dims      JSONB NOT NULL DEFAULT '{}',
    as_of     TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS experiments (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    key         TEXT NOT NULL,
    name        TEXT NOT NULL,
    status      TEXT NOT NULL,
    variants    JSONB NOT NULL,
    primary_kpi TEXT NOT NULL,
    started_at  TIMESTAMPTZ,
    ended_at    TIMESTAMPTZ,
    winner      TEXT NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS experiment_assignments (
    tenant_id     UUID NOT NULL,
    experiment_id UUID NOT NULL,
    subject_id    UUID NOT NULL,
    variant       TEXT NOT NULL,
    assigned_at   TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, experiment_id, subject_id)
);

CREATE TABLE IF NOT EXISTS report_defs (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL,
    query_spec JSONB NOT NULL DEFAULT '{}',
    schedule   TEXT NOT NULL DEFAULT '',
    format     TEXT NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS report_runs (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    report_id    UUID NOT NULL,
    status       TEXT NOT NULL,
    location     TEXT NOT NULL,
    row_count    INT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS observability_signals (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    kind        TEXT NOT NULL,
    service     TEXT NOT NULL,
    name        TEXT NOT NULL,
    value       DOUBLE PRECISION NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT '',
    trace_id    TEXT NOT NULL DEFAULT '',
    attrs       JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_rules (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    name       TEXT NOT NULL,
    metric_key TEXT NOT NULL,
    op         TEXT NOT NULL,
    threshold  DOUBLE PRECISION NOT NULL,
    severity   TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_events (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    rule_id    UUID NOT NULL,
    metric_key TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    severity   TEXT NOT NULL,
    message    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS catalog_assets (
    id             UUID PRIMARY KEY,
    tenant_id      UUID NOT NULL,
    name           TEXT NOT NULL,
    type           TEXT NOT NULL,
    owner          TEXT NOT NULL DEFAULT '',
    description    TEXT NOT NULL DEFAULT '',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    classification TEXT NOT NULL DEFAULT 'internal',
    updated_at     TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS lineage_edges (
    tenant_id UUID NOT NULL,
    from_name TEXT NOT NULL,
    to_name   TEXT NOT NULL,
    kind      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS quality_checks (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    asset_name TEXT NOT NULL,
    check_type TEXT NOT NULL,
    passed     BOOLEAN NOT NULL,
    score      DOUBLE PRECISION NOT NULL,
    details    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
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
