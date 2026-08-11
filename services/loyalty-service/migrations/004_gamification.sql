-- Achievements, missions, streaks, spins, collectibles, AI scores.

CREATE TYPE achievement_rule_type AS ENUM (
    'purchase_count',
    'referral_count',
    'spend_minor',
    'mission_code'
);

CREATE TABLE achievements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    code        TEXT NOT NULL,
    title       TEXT NOT NULL,
    rule_type   achievement_rule_type NOT NULL,
    threshold   BIGINT NOT NULL DEFAULT 0,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_achievements_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_achievements_threshold CHECK (threshold >= 0)
);

CREATE TABLE achievement_unlocks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    account_id      UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    achievement_id  UUID NOT NULL REFERENCES achievements (id),
    code            TEXT NOT NULL,
    unlocked_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_achievement_unlocks UNIQUE (tenant_id, account_id, achievement_id)
);

CREATE TABLE missions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    code           TEXT NOT NULL,
    title          TEXT NOT NULL,
    target_count   BIGINT NOT NULL DEFAULT 1,
    reward_points  BIGINT NOT NULL DEFAULT 0,
    achievement    TEXT NOT NULL DEFAULT '',
    active         BOOLEAN NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_missions_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_missions_target CHECK (target_count > 0)
);

CREATE TYPE mission_status AS ENUM (
    'active',
    'completed',
    'expired'
);

CREATE TABLE mission_progress (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    account_id    UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    mission_id    UUID NOT NULL REFERENCES missions (id),
    progress      BIGINT NOT NULL DEFAULT 0,
    status        mission_status NOT NULL DEFAULT 'active',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,

    CONSTRAINT uq_mission_progress UNIQUE (tenant_id, account_id, mission_id),
    CONSTRAINT chk_mission_progress CHECK (progress >= 0)
);

CREATE TABLE streaks (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    account_id        UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    current_count     INT NOT NULL DEFAULT 0,
    longest_count     INT NOT NULL DEFAULT 0,
    last_active_date  TEXT NOT NULL DEFAULT '',
    broken            BOOLEAN NOT NULL DEFAULT false,
    recovery_used     BOOLEAN NOT NULL DEFAULT false,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_streaks_account UNIQUE (tenant_id, account_id),
    CONSTRAINT chk_streaks_counts CHECK (current_count >= 0 AND longest_count >= 0)
);

CREATE TABLE spin_campaigns (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    code        TEXT NOT NULL,
    title       TEXT NOT NULL,
    prizes      JSONB NOT NULL DEFAULT '[]'::jsonb,
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_spin_campaigns_code UNIQUE (tenant_id, code)
);

CREATE TABLE spins (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    account_id   UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    campaign_id  UUID NOT NULL REFERENCES spin_campaigns (id),
    prize_code   TEXT NOT NULL,
    points_won   BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_spins_points CHECK (points_won >= 0)
);

CREATE TABLE collectibles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    code        TEXT NOT NULL,
    title       TEXT NOT NULL,
    rarity      TEXT NOT NULL DEFAULT 'common',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_collectibles_code UNIQUE (tenant_id, code)
);

CREATE TABLE owned_collectibles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    account_id      UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    collectible_id  UUID NOT NULL REFERENCES collectibles (id),
    acquired_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ai_scores (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    account_id    UUID NOT NULL REFERENCES loyalty_accounts (id) ON DELETE CASCADE,
    principal_id  UUID NOT NULL,
    churn_score   DOUBLE PRECISION NOT NULL DEFAULT 0,
    ltv_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_ai_scores_account UNIQUE (tenant_id, account_id)
);
