-- Sessions: authenticated sessions with idle and absolute expiry.
CREATE TABLE sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id        UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    device_id           UUID REFERENCES devices (id) ON DELETE SET NULL,
    tenant_id           UUID NOT NULL,
    amr                 TEXT[] NOT NULL DEFAULT '{}',
    acr                 TEXT NOT NULL DEFAULT '0',
    ip                  INET,
    user_agent          TEXT NOT NULL DEFAULT '',
    risk_score          REAL NOT NULL DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    idle_expires_at     TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at          TIMESTAMPTZ,
    revoke_reason       TEXT,

    CONSTRAINT chk_sessions_absolute_after_idle CHECK (absolute_expires_at >= idle_expires_at)
);

COMMENT ON TABLE sessions IS 'Active auth sessions; idle + absolute timeouts enforced.';
COMMENT ON COLUMN sessions.amr IS 'Authentication method references (pwd, otp, mfa, etc.).';
COMMENT ON COLUMN sessions.acr IS 'Authentication context class reference.';
COMMENT ON COLUMN sessions.risk_score IS 'Continuous risk score 0–100 for adaptive controls.';
