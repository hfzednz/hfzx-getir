-- Outbox, audit, and indexes.

CREATE TYPE loyalty_outbox_status AS ENUM (
    'pending',
    'published',
    'failed'
);

CREATE TABLE loyalty_outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    account_id    UUID,
    topic         TEXT NOT NULL,
    key           TEXT NOT NULL DEFAULT '',
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    status        loyalty_outbox_status NOT NULL DEFAULT 'pending',
    attempts      INT NOT NULL DEFAULT 0,
    last_error    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ
);

CREATE TABLE loyalty_audit (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    account_id  UUID,
    action      TEXT NOT NULL,
    actor_id    UUID,
    detail      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_loyalty_accounts_tenant ON loyalty_accounts (tenant_id);
CREATE INDEX idx_point_ledger_account ON point_ledger (tenant_id, account_id, created_at DESC);
CREATE INDEX idx_point_ledger_order ON point_ledger (tenant_id, account_id, order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_memberships_tier ON memberships (tenant_id, tier);
CREATE INDEX idx_cashback_grants_status ON cashback_grants (tenant_id, status);
CREATE INDEX idx_referral_events_referrer ON referral_events (tenant_id, referrer_account, status);
CREATE INDEX idx_mission_progress_account ON mission_progress (tenant_id, account_id, status);
CREATE INDEX idx_achievement_unlocks_account ON achievement_unlocks (tenant_id, account_id);
CREATE INDEX idx_spins_account ON spins (tenant_id, account_id, created_at DESC);
CREATE INDEX idx_loyalty_outbox_pending ON loyalty_outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_loyalty_audit_tenant ON loyalty_audit (tenant_id, created_at DESC);
