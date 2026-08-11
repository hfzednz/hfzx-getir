-- Membership tiers + rewards / redemptions.

CREATE TABLE membership_tiers (
    tenant_id         UUID NOT NULL,
    code              TEXT NOT NULL,
    name              TEXT NOT NULL,
    threshold_points  BIGINT NOT NULL DEFAULT 0,
    rank              INT NOT NULL DEFAULT 0,
    benefits          JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (tenant_id, code),
    CONSTRAINT chk_membership_tiers_threshold CHECK (threshold_points >= 0)
);

CREATE TABLE memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    account_id  UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    tier        TEXT NOT NULL DEFAULT 'standard',
    since       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_memberships_account UNIQUE (tenant_id, account_id)
);

CREATE TABLE rewards (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    code          TEXT NOT NULL,
    title         TEXT NOT NULL,
    points_cost   BIGINT NOT NULL DEFAULT 0,
    active        BOOLEAN NOT NULL DEFAULT true,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_rewards_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_rewards_points_cost CHECK (points_cost >= 0)
);

CREATE TYPE loyalty_reward_status AS ENUM (
    'available',
    'unlocked',
    'redeemed'
);

CREATE TABLE redemptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    account_id   UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    reward_id    UUID NOT NULL REFERENCES rewards (id),
    status       loyalty_reward_status NOT NULL DEFAULT 'unlocked',
    points_paid  BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_redemptions_points_paid CHECK (points_paid >= 0)
);
