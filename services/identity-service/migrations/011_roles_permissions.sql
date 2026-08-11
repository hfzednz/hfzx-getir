-- RBAC: roles, permissions, inheritance, scoped bindings, temporary grants.
CREATE TYPE role_kind AS ENUM (
    'platform',
    'tenant',
    'department',
    'custom'
);

CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    name            TEXT NOT NULL,
    kind            role_kind NOT NULL DEFAULT 'custom',
    description     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_roles_tenant_name UNIQUE NULLS NOT DISTINCT (tenant_id, name)
);

CREATE TABLE permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource        TEXT NOT NULL,
    action          TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action),
    CONSTRAINT chk_permissions_resource_action CHECK (
        resource <> '' AND action <> '' AND resource !~ ':' AND action !~ ':'
    )
);

CREATE TABLE role_permissions (
    role_id         UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id   UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE role_parents (
    role_id         UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    parent_role_id  UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (role_id, parent_role_id),
    CONSTRAINT chk_role_parents_no_self CHECK (role_id <> parent_role_id)
);

CREATE TABLE principal_roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    role_id         UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    tenant_id       UUID,
    city_id         UUID,
    warehouse_id    UUID,
    granted_by      UUID REFERENCES principals (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ,

    CONSTRAINT uq_principal_roles_scope UNIQUE NULLS NOT DISTINCT (
        principal_id, role_id, tenant_id, city_id, warehouse_id
    )
);

CREATE TABLE temporary_grants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    permission_id   UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    tenant_id       UUID,
    city_id         UUID,
    warehouse_id    UUID,
    reason          TEXT NOT NULL DEFAULT '',
    granted_by      UUID REFERENCES principals (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT chk_temporary_grants_expires CHECK (expires_at > created_at)
);

COMMENT ON TABLE roles IS 'Named role definitions; tenant_id NULL = platform-scoped.';
COMMENT ON TABLE permissions IS 'resource:action permission atoms (stored as separate columns).';
COMMENT ON TABLE role_parents IS 'Role inheritance edges; child inherits parent permissions.';
COMMENT ON TABLE principal_roles IS 'Role bindings with optional tenant/city/warehouse scope.';
COMMENT ON TABLE temporary_grants IS 'Time-boxed direct permission grants with optional scope.';
