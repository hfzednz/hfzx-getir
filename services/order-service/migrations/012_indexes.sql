-- Performance indexes for OMS write path, timeline, saga sweep, and outbox publisher.

-- Orders
CREATE INDEX idx_orders_tenant_status ON orders (tenant_id, status, created_at DESC);
CREATE INDEX idx_orders_tenant_customer ON orders (tenant_id, customer_principal_id, created_at DESC);
CREATE INDEX idx_orders_tenant_placed ON orders (tenant_id, placed_at DESC)
    WHERE placed_at IS NOT NULL;
CREATE INDEX idx_orders_tenant_type ON orders (tenant_id, type, created_at DESC);
CREATE INDEX idx_orders_scheduled ON orders (tenant_id, scheduled_at)
    WHERE status IN ('draft', 'pending_payment', 'inventory_reservation', 'warehouse_assigned')
      AND scheduled_at IS NOT NULL;
CREATE INDEX idx_orders_parent ON orders (parent_order_id)
    WHERE parent_order_id IS NOT NULL;

-- Order lines
CREATE INDEX idx_order_lines_order ON order_lines (order_id, sort_order);
CREATE INDEX idx_order_lines_variant ON order_lines (tenant_id, variant_id);
CREATE INDEX idx_order_lines_warehouse ON order_lines (warehouse_id)
    WHERE warehouse_id IS NOT NULL;

-- Order events (timeline)
CREATE INDEX idx_order_events_order_time ON order_events (order_id, occurred_at DESC);
CREATE INDEX idx_order_events_tenant_time ON order_events (tenant_id, occurred_at DESC);
CREATE INDEX idx_order_events_type ON order_events (order_id, type);

-- Saga instances
CREATE INDEX idx_saga_instances_order ON saga_instances (order_id, created_at DESC);
CREATE INDEX idx_saga_instances_tenant_status ON saga_instances (tenant_id, status, created_at DESC);
CREATE INDEX idx_saga_instances_running ON saga_instances (status, updated_at)
    WHERE status IN ('pending', 'running', 'compensating');

-- Saga steps
CREATE INDEX idx_saga_steps_saga ON saga_steps (saga_id, created_at);
CREATE INDEX idx_saga_steps_order ON saga_steps (order_id, created_at);
CREATE INDEX idx_saga_steps_pending ON saga_steps (status, updated_at)
    WHERE status IN ('pending', 'failed');

-- Fulfillments
CREATE INDEX idx_fulfillments_order ON fulfillments (order_id);
CREATE INDEX idx_fulfillments_warehouse_status ON fulfillments (warehouse_id, status);
CREATE INDEX idx_fulfillments_tenant_status ON fulfillments (tenant_id, status, created_at DESC);
CREATE INDEX idx_fulfillments_reservation ON fulfillments (reservation_id)
    WHERE reservation_id <> '';
CREATE INDEX idx_fulfillments_ref ON fulfillments (fulfillment_ref)
    WHERE fulfillment_ref <> '';

-- Returns
CREATE INDEX idx_returns_order ON returns (order_id, created_at DESC);
CREATE INDEX idx_returns_tenant_status ON returns (tenant_id, status, created_at DESC);
CREATE INDEX idx_return_lines_return ON return_lines (return_id);
CREATE INDEX idx_return_lines_order_line ON return_lines (order_line_id);

-- Refunds
CREATE INDEX idx_refunds_order ON refunds (order_id, created_at DESC);
CREATE INDEX idx_refunds_tenant_status ON refunds (tenant_id, status, created_at DESC);
CREATE INDEX idx_refunds_return ON refunds (return_id)
    WHERE return_id IS NOT NULL;
CREATE INDEX idx_refunds_payment_ref ON refunds (payment_refund_ref)
    WHERE payment_refund_ref <> '';

-- Outbox publisher sweep
CREATE INDEX idx_outbox_unpublished ON outbox (created_at)
    WHERE published_at IS NULL;
CREATE INDEX idx_outbox_tenant_time ON outbox (tenant_id, created_at DESC);
CREATE INDEX idx_outbox_order ON outbox (order_id)
    WHERE order_id IS NOT NULL;
CREATE INDEX idx_outbox_topic ON outbox (topic, created_at DESC);

-- Modifications audit
CREATE INDEX idx_modifications_order ON modifications (order_id, created_at DESC);
CREATE INDEX idx_modifications_tenant_time ON modifications (tenant_id, created_at DESC);

-- Webhooks
CREATE INDEX idx_webhooks_tenant_active ON webhooks (tenant_id)
    WHERE is_active = TRUE AND disabled_at IS NULL;
