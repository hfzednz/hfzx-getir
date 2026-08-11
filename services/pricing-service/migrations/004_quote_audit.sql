-- Quote audit snapshots (no promotions storage — discounts via PromoClient.Evaluate).
CREATE TABLE quote_audit (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    quote_id    UUID NOT NULL,
    cart_id     UUID,
    simulated   BOOLEAN NOT NULL DEFAULT false,
    request     JSONB NOT NULL DEFAULT '{}'::jsonb,
    response    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE quote_audit IS 'Quote assembly audit; promo discounts come from promotion-service.';
