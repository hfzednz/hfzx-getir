-- OAuth2 / OIDC clients, redirect URIs, and authorization codes.
CREATE TYPE oauth_client_type AS ENUM (
    'confidential',
    'public'
);

CREATE TABLE oauth_clients (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id           TEXT NOT NULL,
    client_secret_hash  TEXT,
    name                TEXT NOT NULL,
    type                oauth_client_type NOT NULL DEFAULT 'confidential',
    tenant_id           UUID,
    grant_types         TEXT[] NOT NULL DEFAULT '{authorization_code,refresh_token}',
    scopes              TEXT[] NOT NULL DEFAULT '{openid}',
    token_endpoint_auth TEXT NOT NULL DEFAULT 'client_secret_basic',
    require_pkce        BOOLEAN NOT NULL DEFAULT FALSE,
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_oauth_clients_client_id UNIQUE (client_id)
);

CREATE TABLE oauth_redirect_uris (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL REFERENCES oauth_clients (id) ON DELETE CASCADE,
    uri             TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_oauth_redirect_uris_client_uri UNIQUE (client_id, uri)
);

CREATE TABLE oauth_authorization_codes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash           TEXT NOT NULL,
    client_id           UUID NOT NULL REFERENCES oauth_clients (id) ON DELETE CASCADE,
    principal_id        UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    session_id          UUID REFERENCES sessions (id) ON DELETE SET NULL,
    redirect_uri        TEXT NOT NULL,
    scopes              TEXT[] NOT NULL DEFAULT '{}',
    code_challenge      TEXT,
    code_challenge_method TEXT,
    nonce               TEXT,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at             TIMESTAMPTZ,

    CONSTRAINT uq_oauth_authorization_codes_hash UNIQUE (code_hash)
);

COMMENT ON TABLE oauth_clients IS 'Registered OAuth2/OIDC clients.';
COMMENT ON TABLE oauth_redirect_uris IS 'Allowed redirect URIs per client.';
COMMENT ON TABLE oauth_authorization_codes IS 'Single-use authorization codes (hashed at rest).';
