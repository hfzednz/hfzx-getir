-- NEXORA erp-service schema (PostgreSQL)
-- Money columns are BIGINT minor units.

CREATE TABLE IF NOT EXISTS erp_companies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  country TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS erp_fiscal_years (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL REFERENCES erp_companies(id),
  label TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  closed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS erp_accounting_periods (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  fiscal_year_id UUID NOT NULL REFERENCES erp_fiscal_years(id),
  label TEXT NOT NULL,
  start_date TIMESTAMPTZ NOT NULL,
  end_date TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('open','closed'))
);

CREATE TABLE IF NOT EXISTS erp_chart_accounts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  account_type TEXT NOT NULL,
  parent_id UUID NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (tenant_id, company_id, code)
);

CREATE TABLE IF NOT EXISTS erp_journal_entries (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  period_id UUID NOT NULL,
  memo TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  status TEXT NOT NULL,
  ledger_ref TEXT NOT NULL DEFAULT '',
  idempotency_key TEXT NULL,
  created_by UUID NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  posted_at TIMESTAMPTZ NULL,
  UNIQUE (tenant_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS erp_journal_lines (
  id BIGSERIAL PRIMARY KEY,
  journal_id UUID NOT NULL REFERENCES erp_journal_entries(id) ON DELETE CASCADE,
  account_code TEXT NOT NULL,
  cost_center TEXT NOT NULL DEFAULT '',
  debit_minor BIGINT NOT NULL DEFAULT 0,
  credit_minor BIGINT NOT NULL DEFAULT 0,
  memo TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS erp_suppliers (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  tax_id TEXT NOT NULL DEFAULT '',
  country TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, company_id, code)
);

CREATE TABLE IF NOT EXISTS erp_purchase_requests (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  requester_id UUID NOT NULL,
  status TEXT NOT NULL,
  currency CHAR(3) NOT NULL,
  total_minor BIGINT NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_purchase_orders (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  pr_id UUID NULL,
  status TEXT NOT NULL,
  currency CHAR(3) NOT NULL,
  total_minor BIGINT NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_goods_receipts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  po_id UUID NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  received_at TIMESTAMPTZ NOT NULL,
  created_by UUID NULL
);

CREATE TABLE IF NOT EXISTS erp_ap_invoices (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  po_id UUID NULL,
  invoice_number TEXT NOT NULL,
  currency CHAR(3) NOT NULL,
  subtotal_minor BIGINT NOT NULL,
  tax_minor BIGINT NOT NULL DEFAULT 0,
  total_minor BIGINT NOT NULL,
  status TEXT NOT NULL,
  match_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  due_date DATE NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  approved_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS erp_ar_invoices (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  customer_ref TEXT NOT NULL,
  invoice_number TEXT NOT NULL,
  currency CHAR(3) NOT NULL,
  total_minor BIGINT NOT NULL,
  status TEXT NOT NULL,
  due_date DATE NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_bank_accounts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  name TEXT NOT NULL,
  iban TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  balance_minor BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS erp_bank_transactions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  bank_account_id UUID NOT NULL REFERENCES erp_bank_accounts(id),
  external_ref TEXT NOT NULL DEFAULT '',
  amount_minor BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  booked_at TIMESTAMPTZ NOT NULL,
  reconciled BOOLEAN NOT NULL DEFAULT FALSE,
  memo TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS erp_budgets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  label TEXT NOT NULL,
  period TEXT NOT NULL,
  year INT NOT NULL,
  status TEXT NOT NULL,
  currency CHAR(3) NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  approved_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS erp_fixed_assets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  cost_minor BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  useful_life_months INT NOT NULL,
  accum_dep_minor BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  acquired_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, company_id, code)
);

CREATE TABLE IF NOT EXISTS erp_tax_returns (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  kind TEXT NOT NULL,
  period_label TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  taxable_minor BIGINT NOT NULL,
  tax_minor BIGINT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_expense_reports (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  employee_id UUID NOT NULL,
  currency CHAR(3) NOT NULL,
  total_minor BIGINT NOT NULL,
  status TEXT NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_approvals (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  kind TEXT NOT NULL,
  subject_id UUID NOT NULL,
  status TEXT NOT NULL,
  steps_json JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  decided_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS erp_payroll_batches (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  currency CHAR(3) NOT NULL,
  total_minor BIGINT NOT NULL,
  status TEXT NOT NULL,
  external_ref TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_outbox (
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

CREATE INDEX IF NOT EXISTS idx_erp_periods_company ON erp_accounting_periods(tenant_id, company_id);
CREATE INDEX IF NOT EXISTS idx_erp_ap_status ON erp_ap_invoices(tenant_id, company_id, status);
CREATE INDEX IF NOT EXISTS idx_erp_approvals_pending ON erp_approvals(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_erp_outbox_pending ON erp_outbox(status) WHERE status = 'pending';
