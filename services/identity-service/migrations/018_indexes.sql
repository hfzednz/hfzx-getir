-- Performance indexes for hot auth / session / RBAC / risk paths.

-- Principals
CREATE INDEX idx_principals_tenant_status ON principals (tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_principals_tenant_kind ON principals (tenant_id, kind)
    WHERE deleted_at IS NULL;

-- Identifiers
CREATE INDEX idx_identifiers_principal ON identifiers (principal_id);
CREATE INDEX idx_identifiers_tenant_type ON identifiers (tenant_id, type);

-- Credentials / password history
CREATE INDEX idx_password_history_principal_created ON password_history (principal_id, created_at DESC);

-- MFA / backup / WebAuthn
CREATE INDEX idx_mfa_factors_principal_type ON mfa_factors (principal_id, type)
    WHERE disabled_at IS NULL;
CREATE INDEX idx_backup_codes_principal_unused ON backup_codes (principal_id)
    WHERE used_at IS NULL;
CREATE INDEX idx_webauthn_credentials_principal ON webauthn_credentials (principal_id)
    WHERE revoked_at IS NULL;

-- Devices
CREATE INDEX idx_devices_principal_last_seen ON devices (principal_id, last_seen_at DESC)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_devices_fingerprint ON devices (fingerprint);

-- Sessions
CREATE INDEX idx_sessions_principal_active ON sessions (principal_id, last_seen_at DESC)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_device ON sessions (device_id)
    WHERE device_id IS NOT NULL AND revoked_at IS NULL;
CREATE INDEX idx_sessions_tenant_created ON sessions (tenant_id, created_at DESC);
CREATE INDEX idx_sessions_idle_expires ON sessions (idle_expires_at)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_absolute_expires ON sessions (absolute_expires_at)
    WHERE revoked_at IS NULL;

-- Refresh tokens
CREATE INDEX idx_refresh_tokens_session ON refresh_tokens (session_id)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family_id)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_principal_expires ON refresh_tokens (principal_id, expires_at)
    WHERE revoked_at IS NULL;

-- RBAC
CREATE INDEX idx_roles_tenant ON roles (tenant_id);
CREATE INDEX idx_role_parents_parent ON role_parents (parent_role_id);
CREATE INDEX idx_principal_roles_principal ON principal_roles (principal_id);
CREATE INDEX idx_principal_roles_role ON principal_roles (role_id);
CREATE INDEX idx_principal_roles_scope ON principal_roles (tenant_id, city_id, warehouse_id);
CREATE INDEX idx_temporary_grants_principal_active ON temporary_grants (principal_id, expires_at)
    WHERE revoked_at IS NULL;

-- OAuth
CREATE INDEX idx_oauth_authorization_codes_client ON oauth_authorization_codes (client_id)
    WHERE used_at IS NULL;
CREATE INDEX idx_oauth_authorization_codes_expires ON oauth_authorization_codes (expires_at)
    WHERE used_at IS NULL;

-- Login attempts / risk / audit (time-series)
CREATE INDEX idx_login_attempts_identifier_created ON login_attempts (identifier, created_at DESC);
CREATE INDEX idx_login_attempts_principal_created ON login_attempts (principal_id, created_at DESC)
    WHERE principal_id IS NOT NULL;
CREATE INDEX idx_login_attempts_ip_created ON login_attempts (ip, created_at DESC)
    WHERE ip IS NOT NULL;
CREATE INDEX idx_risk_events_principal_created ON risk_events (principal_id, created_at DESC)
    WHERE principal_id IS NOT NULL;
CREATE INDEX idx_risk_events_session_created ON risk_events (session_id, created_at DESC)
    WHERE session_id IS NOT NULL;
CREATE INDEX idx_risk_events_type_created ON risk_events (event_type, created_at DESC);
CREATE INDEX idx_audit_events_tenant_created ON audit_events (tenant_id, created_at DESC);
CREATE INDEX idx_audit_events_actor_created ON audit_events (actor_id, created_at DESC)
    WHERE actor_id IS NOT NULL;
CREATE INDEX idx_audit_events_resource ON audit_events (resource_type, resource_id, created_at DESC);
CREATE INDEX idx_audit_events_action_created ON audit_events (action, created_at DESC);

-- Consents / security policies
CREATE INDEX idx_consents_principal ON consents (principal_id);
CREATE INDEX idx_security_policies_tenant_enabled ON security_policies (tenant_id)
    WHERE enabled = TRUE;
