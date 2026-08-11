-- feedback
CREATE TABLE IF NOT EXISTS feedback (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    ticket_id UUID,
    conversation_id UUID,
    kind TEXT NOT NULL,
    score INT NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);
