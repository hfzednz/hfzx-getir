-- Checkout session indexes for tenant lookups, recovery, and admin metrics.
CREATE INDEX idx_checkout_sessions_tenant_status
    ON checkout_sessions (tenant_id, status);

CREATE INDEX idx_checkout_sessions_tenant_principal
    ON checkout_sessions (tenant_id, principal_id);

CREATE INDEX idx_checkout_sessions_tenant_cart
    ON checkout_sessions (tenant_id, cart_id);

CREATE INDEX idx_checkout_sessions_tenant_created
    ON checkout_sessions (tenant_id, created_at DESC);

CREATE UNIQUE INDEX uq_checkout_sessions_recovery
    ON checkout_sessions (recovery_token)
    WHERE recovery_token <> '';

CREATE INDEX idx_checkout_sessions_abandoned
    ON checkout_sessions (tenant_id, abandoned_at DESC)
    WHERE status = 'abandoned';

CREATE INDEX idx_checkout_sessions_order_id
    ON checkout_sessions (tenant_id, order_id)
    WHERE order_id <> '';

CREATE INDEX idx_checkout_events_session
    ON checkout_events (session_id, occurred_at);

CREATE INDEX idx_checkout_outbox_pending
    ON checkout_outbox (status, created_at)
    WHERE status = 'pending';
