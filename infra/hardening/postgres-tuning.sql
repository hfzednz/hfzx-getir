-- Hyperscale Postgres tuning guidance (run with DBA review; additive indexes only)
-- Autovacuum aggressiveness for high-churn tables
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.05;
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.02;
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
ALTER SYSTEM SET work_mem = '32MB';
ALTER SYSTEM SET maintenance_work_mem = '1GB';
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET max_parallel_workers_per_gather = 4;
-- Example partition pattern for high-volume event tables (template)
-- CREATE TABLE events_y2026m08 PARTITION OF events FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
-- SELECT pg_reload_conf();
