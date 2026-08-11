-- kb_versions
CREATE TABLE IF NOT EXISTS kb_versions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    article_id UUID NOT NULL REFERENCES kb_articles(id),
    version INT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
