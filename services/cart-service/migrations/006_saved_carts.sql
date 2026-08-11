-- Saved carts / save-for-later snapshots.
CREATE TABLE saved_carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    principal_id    UUID NOT NULL,
    source_cart_id  UUID REFERENCES carts (id) ON DELETE SET NULL,
    name            TEXT NOT NULL DEFAULT '',
    snapshot        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE saved_carts IS 'Named save-for-later cart snapshots for authenticated principals.';
