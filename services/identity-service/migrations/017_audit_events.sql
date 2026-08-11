-- Audit events: local append-only audit trail (also published to Kafka).
-- No UPDATE/DELETE grants for app role in production; trigger blocks mutations.
CREATE TABLE audit_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    actor_id        UUID,
    actor_kind      TEXT NOT NULL DEFAULT 'principal',
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT,
    outcome         TEXT NOT NULL DEFAULT 'success',
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    session_id      UUID,
    request_id      TEXT,
    details         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION audit_events_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only; % not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_events_no_update
    BEFORE UPDATE ON audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE audit_events_immutable();

CREATE TRIGGER trg_audit_events_no_delete
    BEFORE DELETE ON audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE audit_events_immutable();

COMMENT ON TABLE audit_events IS 'Append-only local audit log; mirrored to identity.audit.events Kafka topic.';
COMMENT ON COLUMN audit_events.action IS 'e.g. principal.locked, session.revoked, role.granted.';
COMMENT ON COLUMN audit_events.outcome IS 'success | failure | denied.';
