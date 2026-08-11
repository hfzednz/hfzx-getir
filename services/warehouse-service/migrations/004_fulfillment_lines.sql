-- Line items on a fulfillment order; variant/location/barcode are opaque.
CREATE TABLE fulfillment_lines (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fulfillment_id      UUID NOT NULL REFERENCES fulfillment_orders (id) ON DELETE CASCADE,
    -- Opaque catalog variant UUID.
    variant_id          UUID NOT NULL,
    sku_code            TEXT NOT NULL DEFAULT '',
    qty                 INT NOT NULL CHECK (qty > 0),
    qty_picked          INT NOT NULL DEFAULT 0 CHECK (qty_picked >= 0),
    qty_packed          INT NOT NULL DEFAULT 0 CHECK (qty_packed >= 0),
    -- Opaque inventory location code (not an FK).
    location_code       TEXT NOT NULL DEFAULT '',
    -- Expected scan target for pick verification.
    barcode_expected    TEXT NOT NULL DEFAULT '',
    barcode             TEXT NOT NULL DEFAULT '',
    expiry_required     BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order          INT NOT NULL DEFAULT 0,
    sequence            INT NOT NULL DEFAULT 0,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_fulfillment_lines_sku CHECK (char_length(sku_code) <= 128),
    CONSTRAINT chk_fulfillment_lines_picked CHECK (qty_picked <= qty),
    CONSTRAINT chk_fulfillment_lines_packed CHECK (qty_packed <= qty_picked)
);

COMMENT ON TABLE fulfillment_lines IS 'Fulfillment line projection; variant_id/location_code/barcode are opaque.';
COMMENT ON COLUMN fulfillment_lines.variant_id IS 'Opaque catalog variant UUID — not an FK.';
COMMENT ON COLUMN fulfillment_lines.location_code IS 'Opaque inventory location code hint for pick route.';
COMMENT ON COLUMN fulfillment_lines.barcode_expected IS 'Expected barcode/QR/RFID for scan validation.';
