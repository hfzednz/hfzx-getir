-- Tax display rules (not a fiscal ledger).
CREATE TABLE tax_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    rate_bps    INT NOT NULL,
    inclusive   BOOLEAN NOT NULL DEFAULT false,
    region_id   UUID,
    active      BOOLEAN NOT NULL DEFAULT true,
    priority    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tax_rules_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_tax_rules_rate CHECK (rate_bps >= 0 AND rate_bps <= 100000)
);

COMMENT ON TABLE tax_rules IS 'Display tax rates in basis points; pricing does not own tax remittance.';
