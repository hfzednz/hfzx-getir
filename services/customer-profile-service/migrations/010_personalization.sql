-- Personalization: homepage/category/recommendation/search/delivery/promotion prefs + habits.
CREATE TABLE personalization (
    profile_id              UUID PRIMARY KEY REFERENCES customer_profiles (id) ON DELETE CASCADE,
    homepage                JSONB NOT NULL DEFAULT '{}'::jsonb,
    category                JSONB NOT NULL DEFAULT '{}'::jsonb,
    recommendation          JSONB NOT NULL DEFAULT '{}'::jsonb,
    search                  JSONB NOT NULL DEFAULT '{}'::jsonb,
    delivery                JSONB NOT NULL DEFAULT '{}'::jsonb,
    promotion               JSONB NOT NULL DEFAULT '{}'::jsonb,
    shopping_habits         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE personalization IS 'Personalization profile vectors/prefs; ranking inference owned by recommendation-service.';
COMMENT ON COLUMN personalization.homepage IS 'Homepage layout / module prefs.';
COMMENT ON COLUMN personalization.category IS 'Category browsing prefs.';
COMMENT ON COLUMN personalization.recommendation IS 'Recommendation opt-outs / affinity hints.';
COMMENT ON COLUMN personalization.search IS 'Search ranking / recent query prefs.';
COMMENT ON COLUMN personalization.delivery IS 'Delivery personalization (preferred slots, couriers).';
COMMENT ON COLUMN personalization.promotion IS 'Promo eligibility / frequency caps prefs.';
COMMENT ON COLUMN personalization.shopping_habits IS 'Observed shopping habit summary.';
