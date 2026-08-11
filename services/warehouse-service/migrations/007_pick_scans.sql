-- Individual scan attempts during a pick session.
CREATE TABLE pick_scans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES pick_sessions (id) ON DELETE CASCADE,
    line_id         UUID NOT NULL REFERENCES fulfillment_lines (id) ON DELETE RESTRICT,
    scanned_code    TEXT NOT NULL,
    qty             INT NOT NULL DEFAULT 1 CHECK (qty > 0),
    ok              BOOLEAN NOT NULL,
    reason          TEXT NOT NULL DEFAULT '',
    scanned_by      UUID,
    scanned_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE pick_scans IS 'Append-friendly scan verification log; expected barcode matched in domain.';
COMMENT ON COLUMN pick_scans.scanned_code IS 'Raw barcode/QR/RFID value from device.';
COMMENT ON COLUMN pick_scans.ok IS 'True when ValidatePickScan passed.';
COMMENT ON COLUMN pick_scans.reason IS 'Failure reason when ok=false (mismatch, overpick, empty, etc.).';
