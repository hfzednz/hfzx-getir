-- Wishlist → cart links (wishlist ownership stays elsewhere; cart stores link only).
CREATE TABLE wishlist_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    cart_id         UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    wishlist_id     UUID NOT NULL, -- opaque external wishlist id
    wishlist_item_id UUID,         -- opaque item within wishlist
    variant_id      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wishlist_links_cart_item UNIQUE (cart_id, wishlist_id, wishlist_item_id)
);

COMMENT ON TABLE wishlist_links IS 'Links wishlist items moved into cart; no wishlist SoT ownership.';
