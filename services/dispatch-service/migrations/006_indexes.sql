CREATE INDEX idx_dispatches_tenant ON dispatches (tenant_id);
CREATE INDEX idx_dispatches_tenant_status ON dispatches (tenant_id, status);
CREATE INDEX idx_dispatches_courier ON dispatches (tenant_id, courier_principal_id);
CREATE INDEX idx_dispatches_order ON dispatches (tenant_id, order_id);

CREATE INDEX idx_dispatch_events_dispatch ON dispatch_events (tenant_id, dispatch_id, created_at);
CREATE INDEX idx_courier_snapshots_tenant ON courier_snapshots (tenant_id, available, on_shift);
CREATE INDEX idx_vehicles_tenant ON vehicles (tenant_id);
CREATE INDEX idx_assignment_attempts_dispatch ON assignment_attempts (tenant_id, dispatch_id);
CREATE INDEX idx_batches_tenant ON batches (tenant_id);

CREATE INDEX idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_tenant_agg ON outbox (tenant_id, aggregate_id);
