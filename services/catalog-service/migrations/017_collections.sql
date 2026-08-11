-- Merchandising collections (manual and smart/rule-based).
CREATE TYPE collection_kind AS ENUM (
    'manual',
    'smart',
    'campaign',
    'featured'
);

CREATE TABLE collections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    kind            collection_kind NOT NULL DEFAULT 'manual',
    description     TEXT NOT NULL DEFAULT '',
    image_url       TEXT NOT NULL DEFAULT '',
    -- Smart collection rules, e.g. {"all":[{"attr":"brand","op":"eq","value":"..."}]}
    rules           JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at       TIMESTAMPTZ,
    ends_at         TIMESTAMPTZ,
    sort_order      INT NOT NULL DEFAULT 0,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_collections_tenant_slug UNIQUE (tenant_id, slug),
    CONSTRAINT chk_collections_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_collections_window CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at >= starts_at)
);

CREATE TABLE collection_products (
    collection_id   UUID NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    sort_order      INT NOT NULL DEFAULT 0,
    pinned          BOOLEAN NOT NULL DEFAULT FALSE,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (collection_id, product_id)
);

COMMENT ON TABLE collections IS 'Merchandising collections; smart rules drive automatic membership.';
COMMENT ON COLUMN collections.kind IS 'manual | smart | campaign | featured.';
COMMENT ON COLUMN collections.rules IS 'Smart filter rules JSON; empty for manual collections.';
COMMENT ON TABLE collection_products IS 'Manual/pinned membership; smart membership may be projected/cached.';
