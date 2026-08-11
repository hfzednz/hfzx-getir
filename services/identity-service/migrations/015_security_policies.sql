-- Security policies: tenant/platform authn and session policy bundles.
CREATE TABLE security_policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    -- Password / credential policy
    password_min_length     INT NOT NULL DEFAULT 12,
    password_require_upper  BOOLEAN NOT NULL DEFAULT TRUE,
    password_require_lower  BOOLEAN NOT NULL DEFAULT TRUE,
    password_require_digit  BOOLEAN NOT NULL DEFAULT TRUE,
    password_require_symbol BOOLEAN NOT NULL DEFAULT TRUE,
    password_history_count  INT NOT NULL DEFAULT 5,
    -- MFA / step-up
    mfa_required            BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_required_above_risk REAL NOT NULL DEFAULT 50 CHECK (mfa_required_above_risk >= 0 AND mfa_required_above_risk <= 100),
    -- Session
    session_idle_seconds        INT NOT NULL DEFAULT 1800,
    session_absolute_seconds    INT NOT NULL DEFAULT 86400,
    refresh_token_seconds       INT NOT NULL DEFAULT 604800,
    max_concurrent_sessions     INT,
    -- Lockout
    max_failed_attempts         INT NOT NULL DEFAULT 5,
    lockout_seconds             INT NOT NULL DEFAULT 900,
    -- Risk
    block_above_risk            REAL NOT NULL DEFAULT 90 CHECK (block_above_risk >= 0 AND block_above_risk <= 100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_security_policies_tenant_name UNIQUE NULLS NOT DISTINCT (tenant_id, name),
    CONSTRAINT chk_security_policies_session_ttl CHECK (
        session_idle_seconds > 0
        AND session_absolute_seconds >= session_idle_seconds
        AND refresh_token_seconds > 0
    )
);

COMMENT ON TABLE security_policies IS 'Authn/session/MFA/lockout policy bundles; tenant_id NULL = platform default.';
