CREATE TABLE courier_snapshots (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    courier_principal_id UUID NOT NULL,
    available            BOOLEAN NOT NULL DEFAULT FALSE,
    lat                  DOUBLE PRECISION NOT NULL DEFAULT 0,
    lng                  DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_load         INT NOT NULL DEFAULT 0,
    max_capacity         INT NOT NULL DEFAULT 1,
    rating               DOUBLE PRECISION NOT NULL DEFAULT 0,
    vehicle_type         vehicle_type NOT NULL DEFAULT 'bike',
    on_shift             BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_courier_snapshot UNIQUE (tenant_id, courier_principal_id),
    CONSTRAINT chk_courier_load CHECK (current_load >= 0 AND max_capacity >= 0)
);

COMMENT ON TABLE courier_snapshots IS 'Courier availability snapshot for assignment (not courier UI).';
