CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE dispatch_status AS ENUM (
    'queued', 'assigned', 'pickup_started', 'picked_up',
    'in_transit', 'arrived', 'delivered', 'failed'
);

CREATE TYPE vehicle_type AS ENUM ('bike', 'scooter', 'car', 'van');
CREATE TYPE pod_type AS ENUM ('otp', 'qr', 'photo', 'signature', 'gps');
CREATE TYPE fail_reason AS ENUM (
    'customer_unavailable', 'address_issue', 'refused',
    'damaged', 'courier_issue', 'other'
);

CREATE TABLE dispatches (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    order_id             UUID NOT NULL,
    fulfillment_id       UUID,
    warehouse_id         UUID,
    courier_principal_id UUID,
    vehicle_id           UUID,
    status               dispatch_status NOT NULL DEFAULT 'queued',
    pickup_lat           DOUBLE PRECISION NOT NULL,
    pickup_lng           DOUBLE PRECISION NOT NULL,
    dropoff_lat          DOUBLE PRECISION NOT NULL,
    dropoff_lng          DOUBLE PRECISION NOT NULL,
    required_vehicle     vehicle_type NOT NULL DEFAULT 'bike',
    batch_id             UUID,
    route_id             UUID,
    eta_seconds          INT,
    pod_type             pod_type,
    pod_reference        TEXT NOT NULL DEFAULT '',
    fail_reason          fail_reason,
    fail_note            TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE dispatches IS 'Delivery dispatch jobs; opaque order/fulfillment/warehouse refs.';
