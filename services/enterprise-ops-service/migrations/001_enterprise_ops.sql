CREATE TABLE IF NOT EXISTS eo_org_nodes (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  parent_id UUID NULL,
  manager_ref TEXT NOT NULL DEFAULT '',
  country_code TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS eo_policies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  title TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  version TEXT NOT NULL,
  body_uri TEXT NOT NULL DEFAULT '',
  owner_ref TEXT NOT NULL DEFAULT '',
  approved_by TEXT NOT NULL DEFAULT '',
  approved_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS eo_portfolios (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  owner_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_programs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  portfolio_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_projects (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  program_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  budget_minor BIGINT NOT NULL DEFAULT 0,
  currency TEXT NOT NULL DEFAULT 'TRY',
  health TEXT NOT NULL DEFAULT 'green',
  owner_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_milestones (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  project_id UUID NOT NULL,
  name TEXT NOT NULL,
  due_at TIMESTAMPTZ NOT NULL,
  done BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_objectives (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  period TEXT NOT NULL,
  title TEXT NOT NULL,
  owner_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_key_results (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  objective_id UUID NOT NULL,
  title TEXT NOT NULL,
  target DOUBLE PRECISION NOT NULL,
  current DOUBLE PRECISION NOT NULL DEFAULT 0,
  unit TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_kpis (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  target DOUBLE PRECISION NOT NULL,
  period TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_risks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  title TEXT NOT NULL,
  category TEXT NOT NULL,
  likelihood INT NOT NULL,
  impact INT NOT NULL,
  score INT NOT NULL,
  status TEXT NOT NULL,
  owner_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_continuity (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  rto_hours INT NOT NULL DEFAULT 0,
  rpo_hours INT NOT NULL DEFAULT 0,
  priority INT NOT NULL DEFAULT 1,
  critical_service TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  activated_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS eo_audits (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  title TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  scheduled_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_findings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  audit_id UUID NOT NULL,
  severity TEXT NOT NULL,
  title TEXT NOT NULL,
  capa TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_meetings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  agenda TEXT NOT NULL DEFAULT '',
  minutes_uri TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_decisions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  meeting_id UUID NULL,
  decided_by TEXT NOT NULL,
  votes_for INT NOT NULL DEFAULT 0,
  votes_against INT NOT NULL DEFAULT 0,
  impact TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_knowledge (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  title TEXT NOT NULL,
  kind TEXT NOT NULL,
  uri TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_resources (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  team_code TEXT NOT NULL,
  period TEXT NOT NULL,
  capacity_fte DOUBLE PRECISION NOT NULL,
  allocated_fte DOUBLE PRECISION NOT NULL,
  utilization DOUBLE PRECISION NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS eo_outbox (
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

CREATE INDEX IF NOT EXISTS idx_eo_org_tenant ON eo_org_nodes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_eo_outbox_pending ON eo_outbox(status) WHERE status = 'pending';
