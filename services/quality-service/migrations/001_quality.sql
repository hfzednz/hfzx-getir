CREATE TABLE IF NOT EXISTS qa_suites (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  owner TEXT NOT NULL DEFAULT '',
  path TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS qa_runs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  suite_key TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  trigger TEXT NOT NULL,
  commit_sha TEXT NOT NULL DEFAULT '',
  branch TEXT NOT NULL DEFAULT '',
  environment TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NOT NULL,
  finished_at TIMESTAMPTZ NULL,
  summary_json JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS qa_results (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  run_id UUID NOT NULL REFERENCES qa_runs(id),
  name TEXT NOT NULL,
  class_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  message TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS qa_coverage (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  run_id UUID NOT NULL,
  service TEXT NOT NULL,
  line_pct DOUBLE PRECISION NOT NULL,
  branch_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
  api_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
  workflow_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS qa_gate_policies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  kind TEXT NOT NULL,
  thresholds JSONB NOT NULL DEFAULT '{}',
  required BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS qa_gate_evals (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  policy_key TEXT NOT NULL,
  run_id UUID NOT NULL,
  passed BOOLEAN NOT NULL,
  score DOUBLE PRECISION NOT NULL,
  details JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS qa_certifications (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  version TEXT NOT NULL,
  commit_sha TEXT NOT NULL,
  status TEXT NOT NULL,
  gate_result_ids UUID[] NOT NULL DEFAULT '{}',
  issued_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NULL,
  notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS qa_flaky (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  test_name TEXT NOT NULL,
  suite_key TEXT NOT NULL,
  fail_count INT NOT NULL DEFAULT 0,
  pass_count INT NOT NULL DEFAULT 0,
  last_status TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, suite_key, test_name)
);

CREATE TABLE IF NOT EXISTS qa_perf (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  run_id UUID NOT NULL,
  scenario TEXT NOT NULL,
  p50_ms DOUBLE PRECISION NOT NULL,
  p95_ms DOUBLE PRECISION NOT NULL,
  p99_ms DOUBLE PRECISION NOT NULL,
  error_rate DOUBLE PRECISION NOT NULL,
  rps DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS qa_security_findings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  run_id UUID NOT NULL,
  tool TEXT NOT NULL,
  severity TEXT NOT NULL,
  title TEXT NOT NULL,
  target TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS qa_outbox (
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

CREATE INDEX IF NOT EXISTS idx_qa_runs_tenant ON qa_runs(tenant_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_qa_outbox_pending ON qa_outbox(status) WHERE status = 'pending';
