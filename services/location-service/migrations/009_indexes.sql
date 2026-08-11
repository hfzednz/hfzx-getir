-- Btree indexes on lat/lng for memory-compatible Postgres.
-- Production: enable PostGIS and add GEOGRAPHY(Point,4326) + GiST, e.g.:
--   ALTER TABLE addresses ADD COLUMN geom geography(Point,4326)
--     GENERATED ALWAYS AS (ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography) STORED;
--   CREATE INDEX idx_addresses_geom ON addresses USING GIST (geom);
-- Same pattern for pois / location_history.

CREATE INDEX idx_addresses_tenant ON addresses (tenant_id);
CREATE INDEX idx_addresses_place ON addresses (tenant_id, place_id);
CREATE INDEX idx_addresses_lat ON addresses (lat);
CREATE INDEX idx_addresses_lng ON addresses (lng);
CREATE INDEX idx_addresses_lat_lng ON addresses (tenant_id, lat, lng);

CREATE INDEX idx_pois_tenant_kind ON pois (tenant_id, kind) WHERE active = TRUE;
CREATE INDEX idx_pois_ref ON pois (tenant_id, ref_id);
CREATE INDEX idx_pois_lat ON pois (lat);
CREATE INDEX idx_pois_lng ON pois (lng);
CREATE INDEX idx_pois_lat_lng ON pois (tenant_id, lat, lng);

CREATE INDEX idx_history_subject ON location_history (tenant_id, subject_type, subject_id, recorded_at DESC);
CREATE INDEX idx_history_lat ON location_history (lat);
CREATE INDEX idx_history_lng ON location_history (lng);

CREATE INDEX idx_geocode_cache_expires ON geocode_cache (expires_at);

CREATE INDEX idx_heat_tenant ON heat_cells (tenant_id);

CREATE INDEX idx_outbox_pending ON outbox (status, created_at) WHERE status = 'pending';
CREATE INDEX idx_outbox_tenant ON outbox (tenant_id);

COMMENT ON INDEX idx_addresses_lat_lng IS 'Btree lat/lng; replace/augment with PostGIS GEOGRAPHY GiST in prod.';
COMMENT ON INDEX idx_pois_lat_lng IS 'Btree lat/lng; replace/augment with PostGIS GEOGRAPHY GiST in prod.';
