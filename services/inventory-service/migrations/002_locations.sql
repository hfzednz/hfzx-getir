-- Location hierarchy inside a warehouse (building → … → bin/container).
CREATE TYPE location_kind AS ENUM (
    'building',
    'floor',
    'zone',
    'aisle',
    'rack',
    'shelf',
    'bin',
    'container'
);

CREATE TYPE zone_type AS ENUM (
    'ambient',
    'cold',
    'frozen',
    'secure'
);

CREATE TABLE locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    parent_id       UUID REFERENCES locations (id) ON DELETE RESTRICT,
    kind            location_kind NOT NULL,
    zone_type       zone_type,
    code            TEXT NOT NULL,
    -- Materialized path of UUIDs, e.g. /root-uuid/child-uuid/self-uuid
    path            TEXT NOT NULL DEFAULT '',
    depth           INT NOT NULL DEFAULT 0 CHECK (depth >= 0),
    name            TEXT NOT NULL DEFAULT '',
    is_pickable     BOOLEAN NOT NULL DEFAULT TRUE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_locations_warehouse_code UNIQUE (warehouse_id, code),
    CONSTRAINT chk_locations_code CHECK (code ~ '^[A-Za-z0-9][A-Za-z0-9_.-]*$'),
    CONSTRAINT chk_locations_no_self_parent CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT chk_locations_zone_type CHECK (
        (kind = 'zone' AND zone_type IS NOT NULL)
        OR (kind <> 'zone' AND zone_type IS NULL)
    )
);

COMMENT ON TABLE locations IS 'Warehouse location tree; path is a materialized UUID path.';
COMMENT ON COLUMN locations.kind IS 'building | floor | zone | aisle | rack | shelf | bin | container.';
COMMENT ON COLUMN locations.zone_type IS 'ambient | cold | frozen | secure; required when kind=zone.';
COMMENT ON COLUMN locations.path IS 'Materialized path /uuid/.../uuid for tree queries without recursive CTE.';
