-- ticket_events
CREATE TABLE IF NOT EXISTS ticket_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    ticket_id UUID NOT NULL REFERENCES tickets(id),
    actor_id UUID,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL
);
