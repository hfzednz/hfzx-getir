-- Category tree with materialized path for fast ancestor/descendant queries.
CREATE TYPE category_kind AS ENUM (
    'department',
    'standard',
    'smart',
    'collection',
    'campaign',
    'featured'
);

CREATE TABLE categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    parent_id       UUID REFERENCES categories (id) ON DELETE RESTRICT,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    kind            category_kind NOT NULL DEFAULT 'standard',
    -- Materialized path of UUIDs, e.g. /root-uuid/child-uuid/self-uuid
    path            TEXT NOT NULL DEFAULT '',
    depth           INT NOT NULL DEFAULT 0 CHECK (depth >= 0),
    sort_order      INT NOT NULL DEFAULT 0,
    description     TEXT NOT NULL DEFAULT '',
    image_url       TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_categories_tenant_slug UNIQUE (tenant_id, slug),
    CONSTRAINT chk_categories_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_categories_no_self_parent CHECK (parent_id IS NULL OR parent_id <> id)
);

COMMENT ON TABLE categories IS 'Hierarchical product taxonomy; path is a materialized UUID path.';
COMMENT ON COLUMN categories.kind IS 'department | standard | smart | collection | campaign | featured.';
COMMENT ON COLUMN categories.path IS 'Materialized path /uuid/.../uuid for tree queries without recursive CTE.';
COMMENT ON COLUMN categories.depth IS '0-based depth from root; root has depth 0.';

CREATE TABLE product_categories (
    product_id      UUID NOT NULL,
    category_id     UUID NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INT NOT NULL DEFAULT 0,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (product_id, category_id)
);

COMMENT ON TABLE product_categories IS 'Many-to-many product membership in categories; FK to products added in 004.';
