CREATE INDEX idx_zones_tenant ON zones (tenant_id);
CREATE INDEX idx_zones_tenant_city ON zones (tenant_id, city);
CREATE INDEX idx_zones_tenant_kind ON zones (tenant_id, kind);
CREATE INDEX idx_zones_tenant_active ON zones (tenant_id, active) WHERE active = TRUE;

CREATE INDEX idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_tenant_agg ON outbox (tenant_id, aggregate_id);
