-- Label metadata + print intents (not courier matching / carrier SoT).
CREATE TYPE label_kind AS ENUM (
    'shipping',
    'qr',
    'barcode',
    'internal',
    'courier',
    'return'
);

CREATE TYPE label_status AS ENUM (
    'draft',
    'ready',
    'printed',
    'void'
);

CREATE TABLE labels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    fulfillment_id  UUID REFERENCES fulfillment_orders (id) ON DELETE SET NULL,
    pack_session_id UUID REFERENCES pack_sessions (id) ON DELETE SET NULL,
    dispatch_unit_id UUID REFERENCES dispatch_units (id) ON DELETE SET NULL,
    kind            label_kind NOT NULL,
    status          label_status NOT NULL DEFAULT 'draft',
    code            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    printer_id      UUID REFERENCES equipment (id) ON DELETE SET NULL,
    printed_at      TIMESTAMPTZ,
    voided_at       TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_labels_warehouse_code UNIQUE (warehouse_id, code),
    CONSTRAINT chk_labels_code CHECK (code <> '')
);

COMMENT ON TABLE labels IS 'Label metadata and print intents for pack/dispatch.';
COMMENT ON COLUMN labels.kind IS 'shipping | qr | barcode | internal | courier | return.';
COMMENT ON COLUMN labels.payload IS 'Render payload / ZPL / template variables.';
