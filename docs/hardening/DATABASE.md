# Database Optimization Guide

- Prefer partial indexes for hot status filters
- Partition high-volume event/outbox tables by month
- Route reads to replicas for list/search; writes to primary
- Autovacuum scale factors per `infra/hardening/postgres-tuning.sql`
- Pool via PgBouncer transaction mode (`postgres-pool.yaml`)
- Never change money representation (int64 minor units)
