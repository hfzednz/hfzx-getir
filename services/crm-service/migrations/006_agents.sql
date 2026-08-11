-- agents
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    display_name TEXT NOT NULL,
    team_id UUID,
    status TEXT NOT NULL DEFAULT 'available',
    skill_ids UUID[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
