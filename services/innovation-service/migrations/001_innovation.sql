CREATE TABLE IF NOT EXISTS inv_modules (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  domain TEXT NOT NULL,
  status TEXT NOT NULL,
  trl INT NOT NULL,
  score DOUBLE PRECISION NOT NULL DEFAULT 0,
  sandbox_only BOOLEAN NOT NULL DEFAULT TRUE,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS inv_experiments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  module_id UUID NOT NULL,
  name TEXT NOT NULL,
  hypothesis TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS inv_simulations (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  params JSONB NOT NULL DEFAULT '{}',
  accuracy DOUBLE PRECISION NOT NULL DEFAULT 0,
  result_summary TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NULL,
  completed_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_twins (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  ref_key TEXT NOT NULL,
  name TEXT NOT NULL,
  model_uri TEXT NOT NULL DEFAULT '',
  version TEXT NOT NULL DEFAULT '1.0.0',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_edge_nodes (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  region TEXT NOT NULL DEFAULT '',
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS inv_iot_devices (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  device_key TEXT NOT NULL,
  kind TEXT NOT NULL,
  location TEXT NOT NULL DEFAULT '',
  connected BOOLEAN NOT NULL DEFAULT FALSE,
  last_seen_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, device_key)
);

CREATE TABLE IF NOT EXISTS inv_robots (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS inv_robot_assignments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  robot_id UUID NOT NULL,
  task_ref TEXT NOT NULL,
  status TEXT NOT NULL,
  assigned_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_drone_missions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  drone_key TEXT NOT NULL,
  order_ref TEXT NOT NULL DEFAULT '',
  landing_zone TEXT NOT NULL,
  status TEXT NOT NULL,
  compliance TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_blockchain_hooks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  purpose TEXT NOT NULL,
  chain_ref TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_xr (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  asset_uri TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_multimodal (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  modes TEXT[] NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_green (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  period TEXT NOT NULL,
  carbon_grams BIGINT NOT NULL DEFAULT 0,
  energy_wh BIGINT NOT NULL DEFAULT 0,
  savings_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, period)
);

CREATE TABLE IF NOT EXISTS inv_quantum (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  adapter TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS inv_outbox (
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

CREATE INDEX IF NOT EXISTS idx_inv_modules_tenant ON inv_modules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inv_outbox_pending ON inv_outbox(status) WHERE status = 'pending';
