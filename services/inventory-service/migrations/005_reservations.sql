-- Reservation headers: soft (cart) / hard (order) holds.
CREATE TYPE reservation_type AS ENUM (
    'soft',
    'hard'
);

CREATE TYPE reservation_status AS ENUM (
    'active',
    'released',
    'consumed',
    'expired'
);

CREATE TABLE reservations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    -- Null = multi-warehouse split; lines carry per-warehouse allocation.
    warehouse_id    UUID REFERENCES warehouses (id) ON DELETE RESTRICT,
    type            reservation_type NOT NULL,
    status          reservation_status NOT NULL DEFAULT 'active',
    expires_at      TIMESTAMPTZ,
    priority        INT NOT NULL DEFAULT 0,
    -- Opaque order/cart/external reference — no order aggregate here.
    external_ref    TEXT NOT NULL DEFAULT '',
    actor_id        UUID,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    released_at     TIMESTAMPTZ,
    consumed_at     TIMESTAMPTZ,

    CONSTRAINT chk_reservations_soft_expires CHECK (
        type = 'hard' OR expires_at IS NOT NULL OR status <> 'active'
    )
);

COMMENT ON TABLE reservations IS 'Soft/hard stock holds; Soft→Hard|Released, Hard→Consumed|Released.';
COMMENT ON COLUMN reservations.external_ref IS 'Opaque cart/order id from upstream; not an FK.';
COMMENT ON COLUMN reservations.warehouse_id IS 'Null when reservation spans multiple warehouses (lines hold WH).';
COMMENT ON COLUMN reservations.type IS 'soft (cart TTL) | hard (committed order hold).';
COMMENT ON COLUMN reservations.status IS 'active | released | consumed | expired.';
