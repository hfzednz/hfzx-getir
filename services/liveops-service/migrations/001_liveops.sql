CREATE TABLE IF NOT EXISTS liveops_flags (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  percentage INT NOT NULL DEFAULT 0,
  rules_json JSONB NOT NULL DEFAULT '[]',
  depends_on TEXT[] NOT NULL DEFAULT '{}',
  version INT NOT NULL,
  emergency_off BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS liveops_configs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  namespace TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}',
  version INT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS liveops_experiments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  kind TEXT NOT NULL,
  hypothesis TEXT NOT NULL DEFAULT '',
  variants_json JSONB NOT NULL DEFAULT '[]',
  primary_metric TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NULL,
  ended_at TIMESTAMPTZ NULL,
  winner TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS liveops_assignments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  experiment_id UUID NOT NULL,
  subject_id TEXT NOT NULL,
  variant_key TEXT NOT NULL,
  assigned_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, experiment_id, subject_id)
);

CREATE TABLE IF NOT EXISTS liveops_events (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  scopes_json JSONB NOT NULL DEFAULT '[]',
  config_patch JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS liveops_changes (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  subject_key TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  decided_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS liveops_rollbacks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  subject_key TEXT NOT NULL,
  from_version INT NOT NULL DEFAULT 0,
  to_version INT NOT NULL DEFAULT 0,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS liveops_outbox (
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

CREATE INDEX IF NOT EXISTS idx_liveops_flags_tenant ON liveops_flags(tenant_id);
CREATE INDEX IF NOT EXISTS idx_liveops_exp_status ON liveops_experiments(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_liveops_outbox_pending ON liveops_outbox(status) WHERE status = 'pending';
