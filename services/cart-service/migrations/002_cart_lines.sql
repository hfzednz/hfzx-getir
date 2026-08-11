-- Cart lines: opaque variant_id, qty with max_qty (available-qty rule at line level).
CREATE TABLE cart_lines (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id             UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    variant_id          UUID NOT NULL, -- opaque catalog variant; no FK
    qty                 INT NOT NULL,
    max_qty             INT NOT NULL DEFAULT 99, -- available qty cap at line level
    notes               TEXT NOT NULL DEFAULT '',
    addons              JSONB NOT NULL DEFAULT '[]'::jsonb,
    replacement_pref    TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_cart_lines_qty CHECK (qty > 0),
    CONSTRAINT chk_cart_lines_max_qty CHECK (max_qty > 0),
    CONSTRAINT chk_cart_lines_qty_lte_max CHECK (qty <= max_qty),
    CONSTRAINT uq_cart_lines_cart_variant UNIQUE (cart_id, variant_id)
);

COMMENT ON COLUMN cart_lines.variant_id IS 'Opaque catalog variant id — cart does not own product content.';
COMMENT ON COLUMN cart_lines.max_qty IS 'Line-level available qty cap (not stock ledger).';
