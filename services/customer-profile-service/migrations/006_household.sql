-- Households: shared family/group profiles with resource sharing flags.
CREATE TABLE households (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    owner_profile_id UUID NOT NULL REFERENCES customer_profiles (id),
    share_addresses BOOLEAN NOT NULL DEFAULT FALSE,
    share_payments  BOOLEAN NOT NULL DEFAULT FALSE,
    share_lists     BOOLEAN NOT NULL DEFAULT FALSE,
    share_wallet    BOOLEAN NOT NULL DEFAULT FALSE,
    share_loyalty   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TYPE household_member_role AS ENUM (
    'owner',
    'adult',
    'child',
    'guest'
);

CREATE TABLE household_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id    UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    role            household_member_role NOT NULL DEFAULT 'adult',
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at         TIMESTAMPTZ,

    CONSTRAINT uq_household_members_household_profile UNIQUE (household_id, profile_id)
);

COMMENT ON TABLE households IS 'Household/group container with sharing flags for addresses, payments, lists, wallet, loyalty.';
COMMENT ON TABLE household_members IS 'Membership of profiles in a household.';
COMMENT ON COLUMN households.owner_profile_id IS 'Primary owner profile; typically role=owner member.';
