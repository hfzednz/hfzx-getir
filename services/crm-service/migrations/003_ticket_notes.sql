-- ticket_notes
CREATE TABLE IF NOT EXISTS ticket_notes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    ticket_id UUID NOT NULL REFERENCES tickets(id),
    author_id UUID NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
