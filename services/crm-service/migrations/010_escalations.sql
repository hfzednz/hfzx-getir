-- escalations
CREATE TABLE IF NOT EXISTS escalations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    ticket_id UUID NOT NULL REFERENCES tickets(id),
    from_priority TEXT NOT NULL,
    to_priority TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    triggered_by_sla BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);
