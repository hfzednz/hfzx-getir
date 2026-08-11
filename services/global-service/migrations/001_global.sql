CREATE TABLE IF NOT EXISTS global_countries (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code CHAR(2) NOT NULL,
  name TEXT NOT NULL,
  default_locale TEXT NOT NULL,
  default_currency CHAR(3) NOT NULL,
  default_tz TEXT NOT NULL,
  status TEXT NOT NULL,
  data_residency TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS global_places (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL REFERENCES global_countries(id),
  parent_id UUID NULL,
  kind TEXT NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_languages (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  rtl BOOLEAN NOT NULL DEFAULT FALSE,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS global_locales (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  locale TEXT NOT NULL,
  language_code TEXT NOT NULL,
  country_code TEXT NOT NULL,
  date_format TEXT NOT NULL,
  time_format TEXT NOT NULL,
  number_format TEXT NOT NULL DEFAULT '',
  currency_format TEXT NOT NULL DEFAULT '',
  first_day_of_week INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, locale)
);

CREATE TABLE IF NOT EXISTS global_translations (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  namespace TEXT NOT NULL,
  key TEXT NOT NULL,
  locale TEXT NOT NULL,
  value TEXT NOT NULL,
  version INT NOT NULL,
  context TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, namespace, key, locale)
);

CREATE TABLE IF NOT EXISTS global_currencies (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  code CHAR(3) NOT NULL,
  name TEXT NOT NULL,
  minor_units INT NOT NULL,
  symbol TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS global_exchange_rates (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  base_currency CHAR(3) NOT NULL,
  quote_currency CHAR(3) NOT NULL,
  rate DOUBLE PRECISION NOT NULL,
  as_of TIMESTAMPTZ NOT NULL,
  source TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_holidays (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  date DATE NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_regional_rules (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  place_id UUID NULL,
  min_order_minor BIGINT NOT NULL DEFAULT 0,
  delivery_fee_minor BIGINT NOT NULL DEFAULT 0,
  currency CHAR(3) NOT NULL,
  legal_age INT NOT NULL DEFAULT 18,
  restricted_skus TEXT[] NOT NULL DEFAULT '{}',
  open_hour TEXT NOT NULL DEFAULT '00:00',
  close_hour TEXT NOT NULL DEFAULT '23:59',
  warehouse_rules JSONB NOT NULL DEFAULT '{}',
  courier_rules JSONB NOT NULL DEFAULT '{}',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, country_id)
);

CREATE TABLE IF NOT EXISTS global_tax_rules (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  place_id UUID NULL,
  kind TEXT NOT NULL,
  rate_bps INT NOT NULL,
  name TEXT NOT NULL,
  exempt_skus TEXT[] NOT NULL DEFAULT '{}',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_privacy (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  framework TEXT NOT NULL,
  consent_required BOOLEAN NOT NULL,
  retention_days INT NOT NULL,
  data_residency TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, country_id)
);

CREATE TABLE IF NOT EXISTS global_payment_availability (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  method_code TEXT NOT NULL,
  enabled BOOLEAN NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_logistics_policy (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  sla_minutes INT NOT NULL,
  holiday_routing BOOLEAN NOT NULL DEFAULT FALSE,
  zone_codes TEXT[] NOT NULL DEFAULT '{}',
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (tenant_id, country_id)
);

CREATE TABLE IF NOT EXISTS global_legal_docs (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  country_id UUID NOT NULL,
  kind TEXT NOT NULL,
  locale TEXT NOT NULL,
  version INT NOT NULL,
  uri TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS global_outbox (
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

CREATE INDEX IF NOT EXISTS idx_global_countries_tenant ON global_countries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_global_translations_ns ON global_translations(tenant_id, namespace, locale);
CREATE INDEX IF NOT EXISTS idx_global_rates_pair ON global_exchange_rates(tenant_id, base_currency, quote_currency, as_of DESC);
CREATE INDEX IF NOT EXISTS idx_global_outbox_pending ON global_outbox(status) WHERE status = 'pending';
