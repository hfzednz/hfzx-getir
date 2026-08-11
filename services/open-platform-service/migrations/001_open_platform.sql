CREATE TABLE IF NOT EXISTS open_apps (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  owner_email TEXT NOT NULL DEFAULT '',
  scopes TEXT[] NOT NULL DEFAULT '{}',
  oauth_client_id TEXT NOT NULL DEFAULT '',
  sandbox BOOLEAN NOT NULL DEFAULT FALSE,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS open_api_keys (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  app_id UUID NOT NULL REFERENCES open_apps(id),
  name TEXT NOT NULL,
  prefix TEXT NOT NULL,
  secret_hash TEXT NOT NULL,
  scopes TEXT[] NOT NULL DEFAULT '{}',
  expires_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL,
  revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS open_catalog (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  title TEXT NOT NULL,
  surface TEXT NOT NULL,
  base_path TEXT NOT NULL,
  service_ref TEXT NOT NULL,
  openapi_path TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS open_api_versions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  catalog_key TEXT NOT NULL,
  version TEXT NOT NULL,
  status TEXT NOT NULL,
  released_at TIMESTAMPTZ NOT NULL,
  deprecated_at TIMESTAMPTZ NULL,
  notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS open_gateway_policies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  route_prefix TEXT NOT NULL,
  target_service TEXT NOT NULL,
  version TEXT NOT NULL,
  rate_limit_rpm INT NOT NULL,
  quota_daily INT NOT NULL DEFAULT 0,
  canary_percent INT NOT NULL DEFAULT 0,
  require_scopes TEXT[] NOT NULL DEFAULT '{}',
  cache_ttl_seconds INT NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS open_webhooks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  app_id UUID NOT NULL,
  url TEXT NOT NULL,
  secret TEXT NOT NULL,
  events TEXT[] NOT NULL DEFAULT '{}',
  version TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS open_webhook_deliveries (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  endpoint_id UUID NOT NULL REFERENCES open_webhooks(id),
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  last_http_status INT NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  signature TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  delivered_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS open_sdks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  language TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  repo_path TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS open_integrations (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  app_id UUID NOT NULL,
  kind TEXT NOT NULL,
  provider TEXT NOT NULL,
  config_ref TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS open_usage (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  app_id UUID NOT NULL,
  day DATE NOT NULL,
  requests BIGINT NOT NULL DEFAULT 0,
  errors BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, app_id, day)
);

CREATE TABLE IF NOT EXISTS open_outbox (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  aggregate_id UUID NOT NULL,
  topic TEXT NOT NULL,
  key TEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_open_apps_tenant ON open_apps(tenant_id);
CREATE INDEX IF NOT EXISTS idx_open_deliveries_pending ON open_webhook_deliveries(status) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_open_outbox_pending ON open_outbox(status) WHERE status = 'pending';
