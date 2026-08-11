-- AI customer model: scores and affinity features (read/recompute; inference elsewhere).
CREATE TABLE ai_customer_models (
    profile_id                  UUID PRIMARY KEY REFERENCES customer_profiles (id) ON DELETE CASCADE,
    frequency                   DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (frequency >= 0),
    avg_order_value_minor       BIGINT NOT NULL DEFAULT 0 CHECK (avg_order_value_minor >= 0),
    churn_prob                  DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (churn_prob >= 0 AND churn_prob <= 1),
    preferred_order_hours       INT[] NOT NULL DEFAULT '{}',
    preferred_order_weekdays    INT[] NOT NULL DEFAULT '{}',
    price_sensitivity           DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (price_sensitivity >= 0 AND price_sensitivity <= 1),
    brand_affinity              JSONB NOT NULL DEFAULT '{}'::jsonb,
    category_affinity           JSONB NOT NULL DEFAULT '{}'::jsonb,
    model_version               TEXT NOT NULL DEFAULT '',
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE ai_customer_models IS 'AI customer scores/affinities; heavy training/inference outside this service.';
COMMENT ON COLUMN ai_customer_models.avg_order_value_minor IS 'Average order value in minor currency units.';
COMMENT ON COLUMN ai_customer_models.churn_prob IS 'Predicted churn probability in [0,1].';
COMMENT ON COLUMN ai_customer_models.preferred_order_hours IS 'Preferred order hours (0-23).';
COMMENT ON COLUMN ai_customer_models.preferred_order_weekdays IS 'Preferred weekdays (0=Sun .. 6=Sat).';
COMMENT ON COLUMN ai_customer_models.brand_affinity IS 'Brand id → score map.';
COMMENT ON COLUMN ai_customer_models.category_affinity IS 'Category id → score map.';
