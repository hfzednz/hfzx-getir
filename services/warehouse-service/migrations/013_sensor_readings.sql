-- Environmental / IoT sensor readings (stub ingest for temp/humidity etc.).
CREATE TYPE sensor_metric AS ENUM (
    'temperature_c',
    'humidity_pct',
    'co2_ppm',
    'door_open',
    'vibration',
    'other'
);

CREATE TABLE sensor_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    equipment_id    UUID REFERENCES equipment (id) ON DELETE SET NULL,
    zone_code       TEXT NOT NULL DEFAULT '',
    metric          sensor_metric NOT NULL,
    value_num       DOUBLE PRECISION,
    value_text      TEXT NOT NULL DEFAULT '',
    unit            TEXT NOT NULL DEFAULT '',
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE sensor_readings IS 'WH sensor ingest stub; not the IoT platform SoT.';
COMMENT ON COLUMN sensor_readings.zone_code IS 'Opaque zone/location hint for cold-chain monitoring.';
COMMENT ON COLUMN sensor_readings.metric IS 'temperature_c | humidity_pct | co2_ppm | door_open | vibration | other.';
