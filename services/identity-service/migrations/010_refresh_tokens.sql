-- Refresh tokens: opaque tokens with rotation families and reuse detection.
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    family_id       UUID NOT NULL,
    rotated_from    UUID REFERENCES refresh_tokens (id) ON DELETE SET NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at      TIMESTAMPTZ,
    revoke_reason   TEXT,

    CONSTRAINT uq_refresh_tokens_hash UNIQUE (token_hash)
);

COMMENT ON TABLE refresh_tokens IS 'Opaque refresh tokens; rotation tracks family for reuse detection.';
COMMENT ON COLUMN refresh_tokens.family_id IS 'Shared across rotations; reuse of old token revokes entire family.';
COMMENT ON COLUMN refresh_tokens.rotated_from IS 'Previous token in the rotation chain.';
