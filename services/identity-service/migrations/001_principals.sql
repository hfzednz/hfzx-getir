-- Principals: canonical identity subjects (users, service accounts, robots, guests).
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE principal_kind AS ENUM (
    'user',
    'service_account',
    'robot',
    'guest'
);

CREATE TYPE principal_status AS ENUM (
    'active',
    'locked',
    'suspended',
    'deleted'
);

CREATE TABLE principals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    kind            principal_kind NOT NULL DEFAULT 'user',
    status          principal_status NOT NULL DEFAULT 'active',
    display_name    TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

COMMENT ON TABLE principals IS 'Canonical identity subjects across tenants.';
COMMENT ON COLUMN principals.kind IS 'Subject type: user | service_account | robot | guest.';
COMMENT ON COLUMN principals.status IS 'Lifecycle: active | locked | suspended | deleted.';
COMMENT ON COLUMN principals.deleted_at IS 'Soft-delete timestamp; status should be deleted when set.';
