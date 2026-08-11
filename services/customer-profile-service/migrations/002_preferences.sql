-- Preferences: one row per customer profile.
CREATE TABLE preferences (
    profile_id              UUID PRIMARY KEY REFERENCES customer_profiles (id) ON DELETE CASCADE,
    favorite_brands         UUID[] NOT NULL DEFAULT '{}',
    favorite_categories     UUID[] NOT NULL DEFAULT '{}',
    favorite_products       UUID[] NOT NULL DEFAULT '{}',
    favorite_stores         UUID[] NOT NULL DEFAULT '{}',
    delivery                JSONB NOT NULL DEFAULT '{}'::jsonb,
    payment                 JSONB NOT NULL DEFAULT '{}'::jsonb,
    notification            JSONB NOT NULL DEFAULT '{}'::jsonb,
    shopping                JSONB NOT NULL DEFAULT '{}'::jsonb,
    theme                   TEXT NOT NULL DEFAULT 'system',
    language                TEXT NOT NULL DEFAULT '',
    accessibility           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE preferences IS 'Per-profile preference bag (brands, delivery, notify, theme).';
COMMENT ON COLUMN preferences.delivery IS 'Delivery prefs: slot windows, door notes defaults, contactless, etc.';
COMMENT ON COLUMN preferences.payment IS 'Preferred payment method hints (refs only, no PAN).';
COMMENT ON COLUMN preferences.notification IS 'Channel/topic notification toggles.';
COMMENT ON COLUMN preferences.shopping IS 'Shopping behaviour prefs (substitutions, eco bag, etc.).';
