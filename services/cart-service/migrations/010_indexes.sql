-- Indexes for cart lookups, merge, abandonment, and outbox drain.
CREATE INDEX idx_carts_tenant_status ON carts (tenant_id, status);
CREATE INDEX idx_carts_tenant_guest ON carts (tenant_id, guest_token) WHERE guest_token <> '';
CREATE INDEX idx_carts_tenant_principal ON carts (tenant_id, principal_id) WHERE principal_id IS NOT NULL;
CREATE INDEX idx_carts_tenant_city ON carts (tenant_id, city_id) WHERE city_id IS NOT NULL;
CREATE INDEX idx_carts_abandoned_at ON carts (tenant_id, abandoned_at) WHERE status = 'abandoned';

CREATE INDEX idx_cart_lines_cart ON cart_lines (cart_id);
CREATE INDEX idx_cart_lines_tenant_variant ON cart_lines (tenant_id, variant_id);

CREATE INDEX idx_cart_coupons_cart ON cart_coupons (cart_id);
CREATE INDEX idx_cart_coupons_code ON cart_coupons (tenant_id, code);

CREATE INDEX idx_cart_quotes_quoted_at ON cart_quotes (tenant_id, quoted_at DESC);

CREATE INDEX idx_cart_reservation_refs_cart ON cart_reservation_refs (cart_id);
CREATE INDEX idx_cart_reservation_refs_active ON cart_reservation_refs (cart_id)
    WHERE released_at IS NULL;

CREATE INDEX idx_saved_carts_principal ON saved_carts (tenant_id, principal_id);

CREATE INDEX idx_wishlist_links_cart ON wishlist_links (cart_id);
CREATE INDEX idx_wishlist_links_wishlist ON wishlist_links (tenant_id, wishlist_id);

CREATE INDEX idx_cart_events_cart_occurred ON cart_events (cart_id, occurred_at);
CREATE INDEX idx_cart_events_tenant_type ON cart_events (tenant_id, type);

CREATE INDEX idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_cart ON outbox (cart_id);
