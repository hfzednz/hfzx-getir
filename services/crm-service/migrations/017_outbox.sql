-- outbox
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    aggregate_id UUID NOT NULL,
    topic TEXT NOT NULL,
    key TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ
);
