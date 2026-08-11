-- delivery events
CREATE TABLE IF NOT EXISTS delivery_events (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    message_id  UUID NOT NULL,
    type        TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
