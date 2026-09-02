-- 002_platform_registry.sql
-- Super-admin tenant/company directory. Normalized tables, not a JSON blob.

CREATE TABLE IF NOT EXISTS platform_companies (
  id TEXT PRIMARY KEY,
  legal_name TEXT NOT NULL,
  trade_name TEXT NOT NULL,
  country_code TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  tenant_count INT NOT NULL DEFAULT 0,
  primary_currency TEXT NOT NULL DEFAULT 'TRY',
  industry TEXT NOT NULL DEFAULT '',
  tax_id TEXT NOT NULL DEFAULT '',
  vat_number TEXT NOT NULL DEFAULT '',
  billing_email TEXT NOT NULL DEFAULT '',
  registered_addr TEXT NOT NULL DEFAULT '',
  default_locale TEXT NOT NULL DEFAULT 'tr-TR',
  time_zone TEXT NOT NULL DEFAULT 'Europe/Istanbul',
  primary_color TEXT NOT NULL DEFAULT '#0B6E6E',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS platform_tenants (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  company_id TEXT NOT NULL DEFAULT '',
  company_name TEXT NOT NULL DEFAULT '',
  isolation_mode TEXT NOT NULL DEFAULT 'shared',
  status TEXT NOT NULL DEFAULT 'pending',
  region TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_platform_tenants_slug UNIQUE (slug)
);

CREATE INDEX IF NOT EXISTS ix_platform_tenants_company ON platform_tenants (company_id);
CREATE INDEX IF NOT EXISTS ix_platform_tenants_status ON platform_tenants (status);

CREATE TABLE IF NOT EXISTS platform_dual_control (
  id TEXT PRIMARY KEY,
  action TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  tenant_name TEXT NOT NULL DEFAULT '',
  requester_id TEXT NOT NULL DEFAULT '',
  requester_email TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_platform_dual_control_tenant ON platform_dual_control (tenant_id, status);

CREATE TABLE IF NOT EXISTS platform_registry_audit (
  id TEXT PRIMARY KEY,
  actor_id TEXT NOT NULL DEFAULT '',
  actor_email TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL,
  resource TEXT NOT NULL,
  resource_id TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  loc TEXT NOT NULL DEFAULT '',
  device TEXT NOT NULL DEFAULT '',
  ip TEXT NOT NULL DEFAULT '',
  session_id TEXT NOT NULL DEFAULT '',
  old_value TEXT,
  new_value TEXT,
  severity TEXT NOT NULL DEFAULT 'info',
  sealed BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS ix_platform_registry_audit_when ON platform_registry_audit (occurred_at DESC);
CREATE INDEX IF NOT EXISTS ix_platform_registry_audit_resource ON platform_registry_audit (resource, resource_id);

CREATE TABLE IF NOT EXISTS platform_people (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL DEFAULT '',
  org_unit_id TEXT NOT NULL DEFAULT '',
  org_unit_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active'
);
