CREATE TABLE IF NOT EXISTS feature_records (
    tenant_id   UUID NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    version     INT NOT NULL,
    values      JSONB NOT NULL DEFAULT '{}',
    tags        JSONB NOT NULL DEFAULT '{}',
    lineage     TEXT NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, entity_type, entity_id, name, version)
);

CREATE TABLE IF NOT EXISTS model_cards (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    key           TEXT NOT NULL,
    name          TEXT NOT NULL,
    framework     TEXT NOT NULL,
    version       TEXT NOT NULL,
    stage         TEXT NOT NULL,
    artifact_uri  TEXT NOT NULL DEFAULT '',
    metrics       JSONB NOT NULL DEFAULT '{}',
    approved_by   UUID,
    approved_at   TIMESTAMPTZ,
    deploy_strat  TEXT NOT NULL DEFAULT 'stable',
    canary_pct    INT NOT NULL DEFAULT 0,
    shadow        BOOLEAN NOT NULL DEFAULT FALSE,
    fallback_key  TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, key, version)
);

CREATE TABLE IF NOT EXISTS prompt_templates (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    key        TEXT NOT NULL,
    locale     TEXT NOT NULL DEFAULT '',
    system     TEXT NOT NULL DEFAULT '',
    user_tpl   TEXT NOT NULL,
    version    INT NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS conversation_memory (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    session_id UUID NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_runs (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    kind       TEXT NOT NULL,
    input      TEXT NOT NULL,
    output     TEXT NOT NULL,
    steps      JSONB NOT NULL DEFAULT '[]',
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS automation_rules (
    id                UUID PRIMARY KEY,
    tenant_id         UUID NOT NULL,
    name              TEXT NOT NULL,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    priority          INT NOT NULL DEFAULT 0,
    conditions        JSONB NOT NULL,
    actions           JSONB NOT NULL,
    require_approval  BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at        TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS automation_runs (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    rule_id    UUID NOT NULL,
    matched    BOOLEAN NOT NULL,
    approved   BOOLEAN NOT NULL DEFAULT FALSE,
    result     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS drift_reports (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    model_key  TEXT NOT NULL,
    metric     TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    threshold  DOUBLE PRECISION NOT NULL,
    severity   TEXT NOT NULL,
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
