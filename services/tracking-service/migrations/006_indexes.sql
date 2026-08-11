-- Indexes for live locations, history, timelines, geofence events, outbox.
CREATE INDEX idx_courier_locations_updated
    ON courier_locations (tenant_id, updated_at DESC);

CREATE INDEX idx_location_history_courier_recorded
    ON location_history (tenant_id, courier_id, recorded_at DESC);

CREATE INDEX idx_delivery_timelines_order
    ON delivery_timelines (tenant_id, order_id, occurred_at);

CREATE INDEX idx_delivery_timelines_courier
    ON delivery_timelines (tenant_id, courier_id, occurred_at)
    WHERE courier_id IS NOT NULL;

CREATE INDEX idx_geofence_events_courier
    ON geofence_events (tenant_id, courier_id, occurred_at DESC);

CREATE INDEX idx_geofence_events_order
    ON geofence_events (tenant_id, order_id, occurred_at DESC)
    WHERE order_id IS NOT NULL;

CREATE INDEX idx_outbox_pending
    ON outbox (status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_tenant
    ON outbox (tenant_id, created_at DESC);
