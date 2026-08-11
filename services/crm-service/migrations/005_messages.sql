-- messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    sender_role TEXT NOT NULL,
    sender_id UUID,
    body TEXT NOT NULL,
    sentiment TEXT,
    created_at TIMESTAMPTZ NOT NULL
);
