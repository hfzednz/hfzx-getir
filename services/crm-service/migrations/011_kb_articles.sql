-- kb_articles
CREATE TABLE IF NOT EXISTS kb_articles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT 'en',
    tags TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, slug)
);
