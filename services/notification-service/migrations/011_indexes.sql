-- indexes
CREATE INDEX IF NOT EXISTS ix_messages_tenant_principal ON messages (tenant_id, principal_id);
CREATE INDEX IF NOT EXISTS ix_messages_tenant_status ON messages (tenant_id, status);
CREATE INDEX IF NOT EXISTS ix_messages_idempotency ON messages (tenant_id, idempotency_key)
    WHERE idempotency_key <> '';
CREATE INDEX IF NOT EXISTS ix_deliveries_message ON deliveries (tenant_id, message_id);
CREATE INDEX IF NOT EXISTS ix_devices_principal ON devices (tenant_id, principal_id) WHERE active;
CREATE INDEX IF NOT EXISTS ix_inbox_principal ON inbox_items (tenant_id, principal_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_schedules_due ON schedules (status, send_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS ix_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS ix_dlq_tenant ON dlq (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_delivery_events_message ON delivery_events (tenant_id, message_id);
CREATE INDEX IF NOT EXISTS ix_consents_principal ON consents (tenant_id, principal_id);
CREATE INDEX IF NOT EXISTS ix_templates_key ON templates (tenant_id, key, channel, locale, status);
