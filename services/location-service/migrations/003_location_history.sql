CREATE TYPE subject_type AS ENUM ('device', 'customer', 'courier', 'warehouse');

CREATE TABLE location_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    subject_type  subject_type NOT NULL,
    subject_id    TEXT NOT NULL,
    lat           DOUBLE PRECISION NOT NULL,
    lng           DOUBLE PRECISION NOT NULL,
    recorded_at   TIMESTAMPTZ NOT NULL,

    CONSTRAINT chk_history_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_history_lng CHECK (lng >= -180 AND lng <= 180),
    CONSTRAINT chk_history_subject CHECK (subject_id <> '')
);

COMMENT ON TABLE location_history IS 'Privacy-scoped history (capped per subject in app); not live courier GPS SoT.';
