-- Login attempts: auth outcome telemetry for lockout and risk.
CREATE TYPE login_attempt_result AS ENUM (
    'success',
    'invalid_credentials',
    'locked',
    'suspended',
    'mfa_required',
    'mfa_failed',
    'blocked_risk',
    'rate_limited'
);

CREATE TABLE login_attempts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    principal_id    UUID REFERENCES principals (id) ON DELETE SET NULL,
    identifier      TEXT NOT NULL DEFAULT '',
    result          login_attempt_result NOT NULL,
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    device_fingerprint TEXT,
    failure_reason  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE login_attempts IS 'Append-oriented login outcome log for lockout and risk signals.';
