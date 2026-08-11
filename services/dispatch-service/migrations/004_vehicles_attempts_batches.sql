CREATE TABLE vehicles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    plate      TEXT NOT NULL,
    type       vehicle_type NOT NULL,
    capacity   INT NOT NULL DEFAULT 1,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_vehicle_plate UNIQUE (tenant_id, plate),
    CONSTRAINT chk_vehicle_capacity CHECK (capacity >= 0)
);

CREATE TABLE assignment_attempts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    dispatch_id          UUID NOT NULL REFERENCES dispatches(id),
    courier_principal_id UUID,
    strategy             TEXT NOT NULL DEFAULT '',
    success              BOOLEAN NOT NULL DEFAULT FALSE,
    distance_m           DOUBLE PRECISION,
    reason               TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE batches (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    label        TEXT NOT NULL DEFAULT '',
    dispatch_ids UUID[] NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
