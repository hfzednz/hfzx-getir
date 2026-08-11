-- Append-only audit trail for inventory mutations (also published to Kafka).
CREATE TABLE inventory_audit_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID,
    actor_id        UUID,
    actor_kind      TEXT NOT NULL DEFAULT 'principal',
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT,
    warehouse_id    UUID,
    variant_id      UUID,
    outcome         TEXT NOT NULL DEFAULT 'success',
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    request_id      TEXT,
    details         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION inventory_audit_events_immutable()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'inventory_audit_events is append-only; % not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_inventory_audit_no_update
    BEFORE UPDATE ON inventory_audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE inventory_audit_events_immutable();

CREATE TRIGGER trg_inventory_audit_no_delete
    BEFORE DELETE ON inventory_audit_events
    FOR EACH ROW
    EXECUTE PROCEDURE inventory_audit_events_immutable();

COMMENT ON TABLE inventory_audit_events IS 'Append-only local audit log for inventory mutations.';
COMMENT ON COLUMN inventory_audit_events.action IS 'e.g. stock.reserved, stock.adjusted, transfer.completed.';
COMMENT ON COLUMN inventory_audit_events.outcome IS 'success | failure | denied.';
COMMENT ON COLUMN inventory_audit_events.variant_id IS 'Opaque catalog variant UUID when applicable.';
