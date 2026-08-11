-- Password history: prevents reuse of recent password hashes.
CREATE TABLE password_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    password_hash   TEXT NOT NULL,
    algorithm       credential_algorithm NOT NULL DEFAULT 'argon2id',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE password_history IS 'Append-only password hash history for reuse prevention.';
