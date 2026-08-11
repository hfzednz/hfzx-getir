-- indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_tenant_status ON campaigns (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_campaigns_tenant_schedule ON campaigns (tenant_id, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_promotions_tenant_campaign ON promotions (tenant_id, campaign_id);
CREATE INDEX IF NOT EXISTS idx_promotions_tenant_type ON promotions (tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_rules_tenant_promo ON rules (tenant_id, promotion_id);
CREATE INDEX IF NOT EXISTS idx_coupons_tenant_promo ON coupons (tenant_id, promotion_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon ON coupon_redemptions (tenant_id, coupon_id);
CREATE INDEX IF NOT EXISTS idx_vouchers_tenant_principal ON vouchers (tenant_id, principal_id);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_voucher ON voucher_redemptions (tenant_id, voucher_id);
CREATE INDEX IF NOT EXISTS idx_usage_tenant_promo ON usage_counters (tenant_id, promotion_id);
CREATE INDEX IF NOT EXISTS idx_simulations_tenant_created ON simulations (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_outbox_tenant ON outbox (tenant_id, created_at DESC);
