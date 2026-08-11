-- tickets
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    assignee_id UUID,
    team_id UUID,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    category TEXT NOT NULL,
    subject TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT,
    merged_into_id UUID,
    first_response_due TIMESTAMPTZ,
    resolve_due TIMESTAMPTZ,
    first_responded_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    sla_breached BOOLEAN NOT NULL DEFAULT FALSE,
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
