-- templates and versions/locales
CREATE TABLE IF NOT EXISTS templates (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    key          TEXT NOT NULL,
    channel      TEXT NOT NULL,
    locale       TEXT NOT NULL DEFAULT 'en',
    version      INT  NOT NULL DEFAULT 1,
    status       TEXT NOT NULL DEFAULT 'draft',
    subject      TEXT NOT NULL DEFAULT '',
    body         TEXT NOT NULL DEFAULT '',
    variant      TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_templates_tenant_key_channel_locale_version
    ON templates (tenant_id, key, channel, locale, version);
