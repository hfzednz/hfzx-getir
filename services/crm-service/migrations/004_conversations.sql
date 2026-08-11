-- conversations
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    agent_id UUID,
    ticket_id UUID,
    status TEXT NOT NULL,
    channel TEXT NOT NULL DEFAULT 'web',
    transferred_from UUID,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
