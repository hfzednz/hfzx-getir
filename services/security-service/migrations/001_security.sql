-- NEXORA security-service schema

CREATE TABLE IF NOT EXISTS sec_policies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  kind TEXT NOT NULL,
  version INT NOT NULL,
  rego TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS sec_policy_decisions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  policy_key TEXT NOT NULL,
  subject TEXT NOT NULL,
  action TEXT NOT NULL,
  resource TEXT NOT NULL DEFAULT '',
  allow BOOLEAN NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  context_json JSONB NOT NULL DEFAULT '{}',
  evaluated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sec_audit_events (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  actor_id TEXT NOT NULL,
  actor_type TEXT NOT NULL,
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL DEFAULT '',
  resource_id TEXT NOT NULL DEFAULT '',
  outcome TEXT NOT NULL,
  ip TEXT NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  metadata_json JSONB NOT NULL DEFAULT '{}',
  hash TEXT NOT NULL,
  prev_hash TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sec_audit_tenant_time ON sec_audit_events(tenant_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS sec_secrets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  vault_path TEXT NOT NULL,
  version INT NOT NULL,
  rotatable BOOLEAN NOT NULL DEFAULT TRUE,
  expires_at TIMESTAMPTZ NULL,
  last_rotated TIMESTAMPTZ NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS sec_threat_alerts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  severity TEXT NOT NULL,
  subject TEXT NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  indicators_json JSONB NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sec_scan_findings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  source TEXT NOT NULL,
  target TEXT NOT NULL DEFAULT '',
  cve TEXT NOT NULL DEFAULT '',
  severity TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  detected_at TIMESTAMPTZ NOT NULL,
  fixed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sec_incidents (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  severity TEXT NOT NULL,
  status TEXT NOT NULL,
  threat_id UUID NULL,
  timeline_json JSONB NOT NULL DEFAULT '[]',
  playbook_key TEXT NOT NULL DEFAULT '',
  assignee TEXT NOT NULL DEFAULT '',
  opened_at TIMESTAMPTZ NOT NULL,
  closed_at TIMESTAMPTZ NULL,
  postmortem TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sec_compliance_controls (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  framework TEXT NOT NULL,
  control_id TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  evidence_ids UUID[] NOT NULL DEFAULT '{}',
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, framework, control_id)
);

CREATE TABLE IF NOT EXISTS sec_compliance_evidence (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  control_id UUID NOT NULL,
  title TEXT NOT NULL,
  uri TEXT NOT NULL DEFAULT '',
  collected_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sec_compliance_audit_runs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  framework TEXT NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  gaps INT NOT NULL,
  status TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sec_data_assets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  classification TEXT NOT NULL,
  pii_tags TEXT[] NOT NULL DEFAULT '{}',
  retention_days INT NOT NULL DEFAULT 0,
  owner TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sec_privacy_requests (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject_ref TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sec_risks (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  category TEXT NOT NULL,
  likelihood INT NOT NULL,
  impact INT NOT NULL,
  score INT NOT NULL,
  status TEXT NOT NULL,
  owner TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sec_access_requests (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  requester_id TEXT NOT NULL,
  role_hint TEXT NOT NULL DEFAULT '',
  resource TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  ttl_minutes INT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at TIMESTAMPTZ NULL,
  expires_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sec_devices (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  device_id TEXT NOT NULL,
  platform TEXT NOT NULL DEFAULT '',
  attested BOOLEAN NOT NULL DEFAULT FALSE,
  rooted BOOLEAN NOT NULL DEFAULT FALSE,
  jailbroken BOOLEAN NOT NULL DEFAULT FALSE,
  tampered BOOLEAN NOT NULL DEFAULT FALSE,
  trust_score DOUBLE PRECISION NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, device_id)
);

CREATE TABLE IF NOT EXISTS sec_ai_events (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  model_key TEXT NOT NULL DEFAULT '',
  prompt_hash TEXT NOT NULL,
  kind TEXT NOT NULL,
  blocked BOOLEAN NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sec_fraud_signals (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  subject TEXT NOT NULL,
  kind TEXT NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  features_json JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sec_outbox (
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

CREATE INDEX IF NOT EXISTS idx_sec_threats_open ON sec_threat_alerts(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_sec_incidents_open ON sec_incidents(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_sec_outbox_pending ON sec_outbox(status) WHERE status = 'pending';
