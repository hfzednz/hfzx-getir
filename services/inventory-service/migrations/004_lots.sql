-- Lots / batches under a stock balance (FEFO/FIFO allocation).
CREATE TYPE lot_status AS ENUM (
    'good',
    'damaged',
    'quarantine',
    'expired'
);

CREATE TABLE stock_lots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    balance_id      UUID NOT NULL REFERENCES stock_balances (id) ON DELETE RESTRICT,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    variant_id      UUID NOT NULL,
    lot_code        TEXT NOT NULL,
    qty             BIGINT NOT NULL DEFAULT 0 CHECK (qty >= 0),
    expiry_date     DATE,
    mfg_date        DATE,
    status          lot_status NOT NULL DEFAULT 'good',
    received_at     TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_lots_balance_code UNIQUE (balance_id, lot_code),
    CONSTRAINT uq_stock_lots_wh_variant_code UNIQUE (warehouse_id, variant_id, lot_code),
    CONSTRAINT chk_stock_lots_code CHECK (lot_code <> ''),
    CONSTRAINT chk_stock_lots_dates CHECK (
        mfg_date IS NULL OR expiry_date IS NULL OR mfg_date <= expiry_date
    )
);

COMMENT ON TABLE stock_lots IS 'Lot/batch quantities; allocate by FEFO (earliest expiry first).';
COMMENT ON COLUMN stock_lots.variant_id IS 'Opaque catalog variant UUID.';
COMMENT ON COLUMN stock_lots.status IS 'good | damaged | quarantine | expired.';
COMMENT ON COLUMN stock_lots.expiry_date IS 'FEFO rank key; null lots sort last.';
