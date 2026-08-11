-- Reservation lines: per-warehouse / variant quantity (optional lot).
CREATE TABLE reservation_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id  UUID NOT NULL REFERENCES reservations (id) ON DELETE CASCADE,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    qty             BIGINT NOT NULL CHECK (qty > 0),
    lot_id          UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    balance_id      UUID REFERENCES stock_balances (id) ON DELETE RESTRICT,
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_reservation_lines_sku CHECK (char_length(sku_code) <= 128)
);

COMMENT ON TABLE reservation_lines IS 'Per-warehouse reservation allocation; variant_id is opaque catalog ref.';
COMMENT ON COLUMN reservation_lines.lot_id IS 'Optional FEFO-pinned lot; null = allocate at pick time.';
COMMENT ON COLUMN reservation_lines.qty IS 'Units held against Available(); never exceeds Available at reserve time.';
