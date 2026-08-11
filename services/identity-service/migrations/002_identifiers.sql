-- Identifiers: login handles bound to a principal within a tenant.
CREATE TYPE identifier_type AS ENUM (
    'email',
    'phone',
    'username',
    'external'
);

CREATE TABLE identifiers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    type            identifier_type NOT NULL,
    value           TEXT NOT NULL,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_identifiers_type_value_tenant UNIQUE (type, value, tenant_id)
);

COMMENT ON TABLE identifiers IS 'Email, phone, username, or external IdP subject per tenant.';
COMMENT ON COLUMN identifiers.value IS 'Normalized identifier value (lowercase email, E.164 phone, etc.).';
COMMENT ON COLUMN identifiers.verified_at IS 'NULL until ownership proven (OTP, magic link, IdP).';
