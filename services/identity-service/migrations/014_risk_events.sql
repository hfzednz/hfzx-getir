-- Risk events: continuous session / auth risk signals.
CREATE TYPE risk_event_severity AS ENUM (
    'info',
    'low',
    'medium',
    'high',
    'critical'
);

CREATE TABLE risk_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID REFERENCES principals (id) ON DELETE SET NULL,
    session_id      UUID REFERENCES sessions (id) ON DELETE SET NULL,
    device_id       UUID REFERENCES devices (id) ON DELETE SET NULL,
    tenant_id       UUID,
    event_type      TEXT NOT NULL,
    severity        risk_event_severity NOT NULL DEFAULT 'info',
    score_delta     REAL NOT NULL DEFAULT 0,
    score_after     REAL,
    ip              INET,
    details         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE risk_events IS 'Risk signal stream; published to Kafka for fraud-service consumption.';
COMMENT ON COLUMN risk_events.event_type IS 'e.g. impossible_travel, new_device, vpn_detected, credential_stuffing.';
COMMENT ON COLUMN risk_events.details IS 'Structured signal payload (geo, ASN, heuristics).';
