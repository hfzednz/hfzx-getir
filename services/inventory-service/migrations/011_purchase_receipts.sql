-- Purchase receipts against opaque PO references (no purchasing aggregate here).
CREATE TYPE purchase_receipt_status AS ENUM (
    'draft',
    'receiving',
    'qc_hold',
    'completed',
    'cancelled'
);

CREATE TYPE qc_status AS ENUM (
    'pending',
    'passed',
    'failed',
    'partial',
    'skipped'
);

CREATE TABLE purchase_receipts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    -- Opaque purchase-order reference from upstream purchasing system.
    po_ref          TEXT NOT NULL,
    status          purchase_receipt_status NOT NULL DEFAULT 'draft',
    qc_status       qc_status NOT NULL DEFAULT 'pending',
    supplier_ref    TEXT NOT NULL DEFAULT '',
    actor_id        UUID,
    notes           TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    received_at     TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,

    CONSTRAINT chk_purchase_receipts_po_ref CHECK (po_ref <> '')
);

CREATE TABLE purchase_receipt_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_id      UUID NOT NULL REFERENCES purchase_receipts (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    lot_code        TEXT NOT NULL DEFAULT '',
    lot_id          UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    qty_expected    BIGINT NOT NULL DEFAULT 0 CHECK (qty_expected >= 0),
    qty_received    BIGINT NOT NULL DEFAULT 0 CHECK (qty_received >= 0),
    expiry_date     DATE,
    mfg_date        DATE,
    qc_status       qc_status NOT NULL DEFAULT 'pending',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_purchase_receipt_lines_sku CHECK (char_length(sku_code) <= 128)
);

COMMENT ON TABLE purchase_receipts IS 'Inbound PO receipts; po_ref is opaque — no PO aggregate owned here.';
COMMENT ON TABLE purchase_receipt_lines IS 'Received quantities per variant; posts purchase_receipt movements.';
COMMENT ON COLUMN purchase_receipts.po_ref IS 'Opaque purchase-order identifier from upstream.';
COMMENT ON COLUMN purchase_receipts.qc_status IS 'pending | passed | failed | partial | skipped.';
