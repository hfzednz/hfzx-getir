-- Dynamic pricing rules (time-of-day / inventory_hint).
CREATE TYPE dynamic_kind AS ENUM ('percent', 'fixed');
CREATE TYPE dynamic_trigger AS ENUM ('time_of_day', 'inventory_hint');

CREATE TABLE dynamic_rules (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    code                 TEXT NOT NULL,
    kind                 dynamic_kind NOT NULL,
    trigger              dynamic_trigger NOT NULL,
    adjustment_bps       INT NOT NULL DEFAULT 0,
    adjustment_minor     BIGINT NOT NULL DEFAULT 0,
    start_hour           INT NOT NULL DEFAULT 0,
    end_hour             INT NOT NULL DEFAULT 24,
    inventory_threshold  INT NOT NULL DEFAULT 0,
    active               BOOLEAN NOT NULL DEFAULT true,
    priority             INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_dynamic_rules_tenant_code UNIQUE (tenant_id, code),
    CONSTRAINT chk_dynamic_rules_hours CHECK (
        start_hour >= 0 AND start_hour <= 23
        AND end_hour >= 1 AND end_hour <= 24
        AND start_hour < end_hour
    )
);

COMMENT ON TABLE dynamic_rules IS 'Percent/fixed adjustments from time_of_day or inventory_hint ports.';
