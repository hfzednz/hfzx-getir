-- Indexes for routes, legs, ETA snapshots, traffic hints, outbox.
CREATE INDEX idx_routes_tenant ON routes (tenant_id);
CREATE INDEX idx_routes_tenant_status ON routes (tenant_id, status);
CREATE INDEX idx_routes_dispatch ON routes (tenant_id, dispatch_id) WHERE dispatch_id IS NOT NULL;
CREATE INDEX idx_routes_courier ON routes (tenant_id, courier_id) WHERE courier_id IS NOT NULL;

CREATE INDEX idx_route_legs_route ON route_legs (tenant_id, route_id);

CREATE INDEX idx_eta_snapshots_route_captured
    ON eta_snapshots (tenant_id, route_id, captured_at DESC);

CREATE INDEX idx_traffic_hints_tenant_valid
    ON traffic_hints (tenant_id, valid_from, valid_until);

CREATE INDEX idx_traffic_hints_region
    ON traffic_hints (tenant_id, region_key);

CREATE INDEX idx_outbox_pending
    ON outbox (status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_tenant
    ON outbox (tenant_id, created_at DESC);
