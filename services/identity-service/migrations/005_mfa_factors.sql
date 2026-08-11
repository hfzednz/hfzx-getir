-- MFA factors: enrolled second factors (TOTP, SMS, email, WebAuthn, push, hardware).
CREATE TYPE mfa_factor_type AS ENUM (
    'totp',
    'sms',
    'email',
    'webauthn',
    'push',
    'hardware'
);

CREATE TABLE mfa_factors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    type            mfa_factor_type NOT NULL,
    label           TEXT NOT NULL DEFAULT '',
    secret_enc      BYTEA,
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    disabled_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE mfa_factors IS 'Enrolled MFA factors; secret_enc is KMS/app-encrypted material.';
COMMENT ON COLUMN mfa_factors.secret_enc IS 'Encrypted TOTP seed or factor secret; NULL for some types.';
COMMENT ON COLUMN mfa_factors.verified IS 'True after successful enrollment challenge.';
