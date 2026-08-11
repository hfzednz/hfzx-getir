CREATE TABLE dispatch_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    dispatch_id UUID NOT NULL REFERENCES dispatches(id),
    type        TEXT NOT NULL,
    from_status dispatch_status,
    to_status   dispatch_status,
    payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE dispatch_events IS 'Append-only dispatch lifecycle audit.';
