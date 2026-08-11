-- Loyalty / wallet projections only — NOT ledgers (owned by loyalty-service / wallet-service).
CREATE TABLE loyalty_projections (
    profile_id          UUID PRIMARY KEY REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    points              BIGINT NOT NULL DEFAULT 0,
    level               TEXT NOT NULL DEFAULT '',
    tier_code           TEXT NOT NULL DEFAULT '',
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_updated_at   TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallet_projections (
    profile_id          UUID PRIMARY KEY REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    balance_minor       BIGINT NOT NULL DEFAULT 0,
    currency            CHAR(3) NOT NULL DEFAULT 'TRY',
    status              TEXT NOT NULL DEFAULT 'active',
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_updated_at   TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE loyalty_projections IS 'Read-model of loyalty points/level; ledger lives in loyalty-service.';
COMMENT ON TABLE wallet_projections IS 'Read-model of wallet balances; ledger lives in wallet-service.';
COMMENT ON COLUMN loyalty_projections.metadata IS 'Opaque projection metadata from loyalty-service events.';
COMMENT ON COLUMN wallet_projections.balance_minor IS 'Available balance in minor currency units.';
COMMENT ON COLUMN wallet_projections.metadata IS 'Opaque projection metadata from wallet-service events.';
