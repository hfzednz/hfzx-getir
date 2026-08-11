-- Align outbox + fulfillments with domain ports (status lifecycle, line_ids).

CREATE TYPE outbox_status AS ENUM (
    'pending',
    'published',
    'failed'
);

ALTER TABLE outbox
    ADD COLUMN IF NOT EXISTS status outbox_status NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Backfill status from published_at for existing rows.
UPDATE outbox
SET status = 'published',
    updated_at = COALESCE(published_at, updated_at, created_at)
WHERE published_at IS NOT NULL
  AND status = 'pending';

CREATE INDEX IF NOT EXISTS idx_outbox_pending_status
    ON outbox (status, created_at)
    WHERE status = 'pending';

ALTER TABLE fulfillments
    ADD COLUMN IF NOT EXISTS line_ids UUID[] NOT NULL DEFAULT '{}';

COMMENT ON COLUMN outbox.status IS 'pending | published | failed — swept by outbox publisher.';
COMMENT ON COLUMN outbox.updated_at IS 'Last status/attempt update time.';
COMMENT ON COLUMN fulfillments.line_ids IS 'Opaque order_line ids covered by this fulfillment split.';
