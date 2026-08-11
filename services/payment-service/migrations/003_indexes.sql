-- Payment indexes for tenant lookups, idempotency, admin list, outbox drain.
CREATE INDEX idx_payment_intents_tenant_status
    ON payment_intents (tenant_id, status);

CREATE INDEX idx_payment_intents_tenant_principal
    ON payment_intents (tenant_id, principal_id);

CREATE INDEX idx_payment_intents_tenant_order
    ON payment_intents (tenant_id, order_id)
    WHERE order_id <> '';

CREATE INDEX idx_payment_intents_tenant_created
    ON payment_intents (tenant_id, created_at DESC);

CREATE INDEX idx_payment_attempts_intent
    ON payment_attempts (intent_id, created_at);

CREATE INDEX idx_payment_methods_principal
    ON payment_methods (tenant_id, principal_id)
    WHERE active = true;

CREATE INDEX idx_refunds_intent
    ON refunds (intent_id, created_at);

CREATE INDEX idx_chargebacks_tenant
    ON chargebacks (tenant_id, created_at DESC);

CREATE INDEX idx_provider_routes_lookup
    ON provider_routes (tenant_id, method_type, currency)
    WHERE active = true;

CREATE INDEX idx_fraud_scores_intent
    ON fraud_scores (intent_id, created_at DESC);

CREATE INDEX idx_payment_outbox_pending
    ON payment_outbox (status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_payment_audit_tenant
    ON payment_audit (tenant_id, created_at DESC);

CREATE INDEX idx_payment_audit_intent
    ON payment_audit (intent_id, created_at)
    WHERE intent_id IS NOT NULL;
