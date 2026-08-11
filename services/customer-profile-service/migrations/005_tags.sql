-- Tags: tenant-scoped definitions and profile assignments.
CREATE TYPE tag_kind AS ENUM (
    'vip',
    'premium',
    'new',
    'returning',
    'high_value',
    'inactive',
    'risk',
    'fraud_watch',
    'custom'
);

CREATE TABLE tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    kind            tag_kind NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    color           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tags_tenant_name UNIQUE (tenant_id, name)
);

CREATE TABLE profile_tags (
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tag_id          UUID NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    assigned_by     UUID,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    note            TEXT NOT NULL DEFAULT '',

    PRIMARY KEY (profile_id, tag_id)
);

COMMENT ON TABLE tags IS 'Tag catalog per tenant (vip, premium, risk, custom, …).';
COMMENT ON TABLE profile_tags IS 'Many-to-many assignment of tags to profiles.';
COMMENT ON COLUMN profile_tags.assigned_by IS 'Actor principal_id who assigned the tag (nullable for system).';
