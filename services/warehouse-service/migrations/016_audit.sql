-- Append-only audit trail for warehouse mutations (also published to Kafka).
CREATE TABLE warehouse_audit_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    actor_id        UUID,
    actor_kind      TEXT NOT NULL DEFAULT 'principal',
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT,
    warehouse_id    UUID,
    fulfillment_id  UUID,
    task_id         UUID,
    outcome         TEXT NOT NULL DEFAULT 'success',
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    request_id      TEXT,
    details         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION warehouse_audit_events_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'warehouse_audit_events is append-only; % not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_warehouse_audit_no_update
    BEFORE UPDATE ON warehouse_audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE warehouse_audit_events_immutable();

CREATE TRIGGER trg_warehouse_audit_no_delete
    BEFORE DELETE ON warehouse_audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE warehouse_audit_events_immutable();

COMMENT ON TABLE warehouse_audit_events IS 'Append-only local audit log for warehouse ops mutations.';
COMMENT ON COLUMN warehouse_audit_events.action IS 'e.g. fulfillment.received, task.claimed, pick.scan, dispatch.handoff.';
COMMENT ON COLUMN warehouse_audit_events.outcome IS 'success | failure | denied.';
