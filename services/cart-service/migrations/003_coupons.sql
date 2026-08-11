-- Coupons applied on cart (preview codes; pricing owns discount SoT).
CREATE TABLE cart_coupons (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id          UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    tenant_id        UUID NOT NULL,
    code             TEXT NOT NULL,
    discount_minor   BIGINT NOT NULL DEFAULT 0,
    applied_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata         JSONB NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT uq_cart_coupons_cart_code UNIQUE (cart_id, code),
    CONSTRAINT chk_cart_coupons_code CHECK (code <> ''),
    CONSTRAINT chk_cart_coupons_discount_nonneg CHECK (discount_minor >= 0)
);

COMMENT ON TABLE cart_coupons IS 'Preview coupon codes on cart; PricingClient computes real discount.';
