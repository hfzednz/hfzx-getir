CREATE TABLE IF NOT EXISTS platform_deployments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  service TEXT NOT NULL,
  environment TEXT NOT NULL,
  strategy TEXT NOT NULL,
  image_tag TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS platform_scaling_events (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  service TEXT NOT NULL,
  environment TEXT NOT NULL,
  from_replicas INT NOT NULL,
  to_replicas INT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS platform_backups (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  target TEXT NOT NULL,
  status TEXT NOT NULL,
  location TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS platform_recoveries (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS platform_alerts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  severity TEXT NOT NULL,
  status TEXT NOT NULL,
  labels_json JSONB NOT NULL DEFAULT '{}',
  fired_at TIMESTAMPTZ NOT NULL,
  resolved_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS platform_costs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  environment TEXT NOT NULL,
  amount_minor BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  period TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS platform_slo_reports (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  service TEXT NOT NULL,
  objective DOUBLE PRECISION NOT NULL,
  actual DOUBLE PRECISION NOT NULL,
  budget_left DOUBLE PRECISION NOT NULL,
  "window" TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS platform_outbox (
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
