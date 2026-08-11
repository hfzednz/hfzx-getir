-- WebAuthn / FIDO2 credentials (passkeys, security keys).
CREATE TABLE webauthn_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    credential_id   BYTEA NOT NULL,
    public_key      BYTEA NOT NULL,
    aaguid          BYTEA,
    sign_count      BIGINT NOT NULL DEFAULT 0,
    transports      TEXT[] NOT NULL DEFAULT '{}',
    nickname        TEXT NOT NULL DEFAULT '',
    backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    backup_state    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at    TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT uq_webauthn_credential_id UNIQUE (credential_id)
);

COMMENT ON TABLE webauthn_credentials IS 'Registered WebAuthn credentials for passkey / security-key auth.';
COMMENT ON COLUMN webauthn_credentials.sign_count IS 'Signature counter for clone detection.';
COMMENT ON COLUMN webauthn_credentials.transports IS 'Hint transports: usb, nfc, ble, internal, hybrid.';
