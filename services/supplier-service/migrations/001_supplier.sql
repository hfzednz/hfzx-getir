CREATE TABLE IF NOT EXISTS supplier_masters (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  code TEXT NOT NULL,
  legal_name TEXT NOT NULL,
  display_name TEXT NOT NULL DEFAULT '',
  country TEXT NOT NULL,
  tax_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  partner_kinds TEXT[] NOT NULL DEFAULT '{}',
  contacts_json JSONB NOT NULL DEFAULT '[]',
  banking_ref TEXT NOT NULL DEFAULT '',
  erp_supplier_id TEXT NOT NULL DEFAULT '',
  risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  approved_at TIMESTAMPTZ NULL,
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS supplier_documents (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  uri TEXT NOT NULL,
  expires_at TIMESTAMPTZ NULL,
  verified BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_certifications (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  name TEXT NOT NULL,
  issuer TEXT NOT NULL DEFAULT '',
  valid_until TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_contracts (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  title TEXT NOT NULL,
  version INT NOT NULL,
  status TEXT NOT NULL,
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  terms_uri TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_rfqs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  number TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  due_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, number)
);

CREATE TABLE IF NOT EXISTS supplier_quotations (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  rfq_id UUID NOT NULL REFERENCES supplier_rfqs(id),
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  currency TEXT NOT NULL,
  total_minor BIGINT NOT NULL,
  lead_time_days INT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_sourcing_pos (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  company_id UUID NOT NULL,
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  number TEXT NOT NULL,
  status TEXT NOT NULL,
  currency TEXT NOT NULL,
  total_minor BIGINT NOT NULL,
  lines_json JSONB NOT NULL DEFAULT '[]',
  quotation_id UUID NULL,
  erp_po_id TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, number)
);

CREATE TABLE IF NOT EXISTS supplier_shipments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  po_id UUID NOT NULL REFERENCES supplier_sourcing_pos(id),
  asn_number TEXT NOT NULL,
  status TEXT NOT NULL,
  tracking_ref TEXT NOT NULL DEFAULT '',
  warehouse_id TEXT NOT NULL DEFAULT '',
  qc_passed BOOLEAN NULL,
  created_at TIMESTAMPTZ NOT NULL,
  received_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS supplier_invoice_matches (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  po_id UUID NOT NULL,
  invoice_ref TEXT NOT NULL,
  amount_minor BIGINT NOT NULL,
  currency TEXT NOT NULL,
  matched BOOLEAN NOT NULL,
  match_score DOUBLE PRECISION NOT NULL,
  erp_invoice_id TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS marketplace_sellers (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL REFERENCES supplier_masters(id),
  store_name TEXT NOT NULL,
  status TEXT NOT NULL,
  rating_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
  rating_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS marketplace_listings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  seller_id UUID NOT NULL REFERENCES marketplace_sellers(id),
  external_sku TEXT NOT NULL,
  catalog_sku TEXT NOT NULL DEFAULT '',
  price_minor BIGINT NOT NULL,
  currency TEXT NOT NULL,
  stock_hint BIGINT NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS catalog_submissions (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  sku TEXT NOT NULL,
  title TEXT NOT NULL,
  attributes JSONB NOT NULL DEFAULT '{}',
  media_uris TEXT[] NOT NULL DEFAULT '{}',
  version INT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS edi_documents (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  doc_type TEXT NOT NULL,
  direction TEXT NOT NULL,
  payload TEXT NOT NULL,
  mapped_ref TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_scorecards (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  period TEXT NOT NULL,
  delivery_score DOUBLE PRECISION NOT NULL,
  quality_score DOUBLE PRECISION NOT NULL,
  price_score DOUBLE PRECISION NOT NULL,
  lead_time_days_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
  fill_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
  compliance_score DOUBLE PRECISION NOT NULL,
  risk_score DOUBLE PRECISION NOT NULL,
  overall DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_threads (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  supplier_id UUID NOT NULL,
  subject TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_messages (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  thread_id UUID NOT NULL REFERENCES supplier_threads(id),
  sender TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_changes (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  kind TEXT NOT NULL,
  subject_key TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  decided_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS supplier_outbox (
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

CREATE INDEX IF NOT EXISTS idx_supplier_tenant ON supplier_masters(tenant_id);
CREATE INDEX IF NOT EXISTS idx_supplier_po_supplier ON supplier_sourcing_pos(tenant_id, supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_outbox_pending ON supplier_outbox(status) WHERE status = 'pending';
