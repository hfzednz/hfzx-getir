-- Consents: privacy / terms / marketing consent records per principal.
CREATE TABLE consents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    purpose         TEXT NOT NULL,
    version         TEXT NOT NULL,
    granted         BOOLEAN NOT NULL,
    granted_at      TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_consents_principal_purpose_version UNIQUE (principal_id, purpose, version)
);

COMMENT ON TABLE consents IS 'Consent ledger for privacy/terms purposes with versioning.';
COMMENT ON COLUMN consents.purpose IS 'e.g. terms_of_service, privacy_policy, marketing_email.';
