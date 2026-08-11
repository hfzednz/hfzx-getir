-- Product-to-product relations (merchandising / AI suggestions).
CREATE TYPE relation_type AS ENUM (
    'related',
    'alternative',
    'accessory',
    'replacement',
    'complementary',
    'ai'
);

CREATE TABLE product_relations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    source_product_id   UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    target_product_id   UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    type                relation_type NOT NULL DEFAULT 'related',
    sort_order          INT NOT NULL DEFAULT 0,
    score               DOUBLE PRECISION,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_relations UNIQUE (source_product_id, target_product_id, type),
    CONSTRAINT chk_product_relations_no_self CHECK (source_product_id <> target_product_id)
);

COMMENT ON TABLE product_relations IS 'Directed product relations for merchandising and AI.';
COMMENT ON COLUMN product_relations.type IS 'related | alternative | accessory | replacement | complementary | ai.';
COMMENT ON COLUMN product_relations.score IS 'Optional relevance/confidence score (AI or merchandiser).';
