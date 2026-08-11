CREATE TABLE IF NOT EXISTS sa_modules (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  publisher_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  latest_version TEXT NOT NULL DEFAULT '0.1.0',
  icon_uri TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS sa_manifests (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  module_id UUID NOT NULL REFERENCES sa_modules(id),
  version TEXT NOT NULL,
  entry_point TEXT NOT NULL,
  min_shell_version TEXT NOT NULL DEFAULT '1.0.0',
  permissions TEXT[] NOT NULL DEFAULT '{}',
  hooks TEXT[] NOT NULL DEFAULT '{}',
  signature TEXT NOT NULL,
  checksum TEXT NOT NULL,
  bundle_uri TEXT NOT NULL DEFAULT '',
  compatible BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, module_id, version)
);

CREATE TABLE IF NOT EXISTS sa_installs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  module_id UUID NOT NULL REFERENCES sa_modules(id),
  version TEXT NOT NULL,
  status TEXT NOT NULL,
  previous_version TEXT NOT NULL DEFAULT '',
  installed_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, subject_id, module_id)
);

CREATE TABLE IF NOT EXISTS sa_permission_grants (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  module_id UUID NOT NULL,
  permission TEXT NOT NULL,
  granted BOOLEAN NOT NULL DEFAULT TRUE,
  granted_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, subject_id, module_id, permission)
);

CREATE TABLE IF NOT EXISTS sa_store_listings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  module_id UUID NOT NULL REFERENCES sa_modules(id),
  featured BOOLEAN NOT NULL DEFAULT FALSE,
  price_minor BIGINT NOT NULL DEFAULT 0,
  currency TEXT NOT NULL DEFAULT 'TRY',
  subscription BOOLEAN NOT NULL DEFAULT FALSE,
  rating_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
  rating_count INT NOT NULL DEFAULT 0,
  installs BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, module_id)
);

CREATE TABLE IF NOT EXISTS sa_store_ratings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  module_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  score INT NOT NULL CHECK (score BETWEEN 1 AND 5),
  comment TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sa_widgets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  module_id UUID NOT NULL,
  slot TEXT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sa_monetization (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  module_id UUID NOT NULL,
  commission_bps INT NOT NULL DEFAULT 0,
  partner_share_bps INT NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, module_id)
);

CREATE TABLE IF NOT EXISTS sa_launches (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  module_id UUID NOT NULL,
  kind TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sa_outbox (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  aggregate_id UUID NOT NULL,
  topic TEXT NOT NULL,
  key TEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_sa_modules_tenant ON sa_modules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sa_installs_subject ON sa_installs(tenant_id, subject_id);
CREATE INDEX IF NOT EXISTS idx_sa_outbox_pending ON sa_outbox(status) WHERE status = 'pending';
