-- sla_policies
CREATE TABLE IF NOT EXISTS sla_policies (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    priority TEXT NOT NULL,
    first_response_minutes INT NOT NULL,
    resolve_minutes INT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, priority)
);
