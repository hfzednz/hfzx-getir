-- Soft reservation refs from InventoryClient (opaque; no stock ledger here).
CREATE TABLE cart_reservation_refs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id             UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    reservation_ref     TEXT NOT NULL, -- opaque inventory soft-reserve id
    idempotency_key     TEXT NOT NULL DEFAULT '',
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    released_at         TIMESTAMPTZ,

    CONSTRAINT uq_cart_reservation_refs_ref UNIQUE (cart_id, reservation_ref),
    CONSTRAINT chk_cart_reservation_refs_ref CHECK (reservation_ref <> '')
);

COMMENT ON COLUMN cart_reservation_refs.reservation_ref IS 'Opaque InventoryClient soft-reserve ref; no FK across services.';
