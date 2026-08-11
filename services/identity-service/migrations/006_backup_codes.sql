-- Backup codes: one-time MFA recovery codes (hashed at rest).
CREATE TABLE backup_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    code_hash       TEXT NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_backup_codes_principal_hash UNIQUE (principal_id, code_hash)
);

COMMENT ON TABLE backup_codes IS 'Hashed one-time MFA backup codes; used_at set on consumption.';
