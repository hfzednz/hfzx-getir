-- Credentials: password material for principals (one active password per principal).
CREATE TYPE credential_algorithm AS ENUM (
    'argon2id'
);

CREATE TABLE credentials (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id        UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    password_hash       TEXT NOT NULL,
    algorithm           credential_algorithm NOT NULL DEFAULT 'argon2id',
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_credentials_principal UNIQUE (principal_id)
);

COMMENT ON TABLE credentials IS 'Primary password credential; algorithm is argon2id only.';
COMMENT ON COLUMN credentials.password_hash IS 'Encoded argon2id hash (PHC string format).';
