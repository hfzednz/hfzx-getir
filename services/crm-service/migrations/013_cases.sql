-- cases
CREATE TABLE IF NOT EXISTS cases (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    ticket_id UUID,
    case_type TEXT NOT NULL,
    status TEXT NOT NULL,
    title TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    assignee_id UUID,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
