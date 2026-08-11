-- Auth challenges + oauth principal linkage for production identity adapters.
ALTER TABLE oauth_clients
    ADD COLUMN IF NOT EXISTS principal_id UUID REFERENCES principals (id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS otp_challenges (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    phone           TEXT NOT NULL,
    code_hash       TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    attempts        INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_otp_challenges_tenant_phone ON otp_challenges (tenant_id, phone);

CREATE TABLE IF NOT EXISTS magic_link_challenges (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS password_reset_challenges (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mfa_challenges (
    id              UUID PRIMARY KEY,
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    session_hint    UUID NOT NULL,
    factor_type     TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webauthn_ceremonies (
    id              TEXT PRIMARY KEY,
    user_id         BYTEA NOT NULL,
    challenge       BYTEA NOT NULL,
    type            TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
