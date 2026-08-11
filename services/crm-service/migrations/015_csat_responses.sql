-- csat_responses
CREATE TABLE IF NOT EXISTS csat_responses (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    ticket_id UUID,
    conversation_id UUID,
    score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);
