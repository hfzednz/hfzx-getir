# Database Release

## Ordering

1. Take backup / PITR bookmark (label: `pre-release-<tag>`).
2. Apply **expand** migrations only (additive columns/indexes/tables).
3. Deploy services that read/write both old and new shapes.
4. Backfill jobs (idempotent) if required.
5. Later release: **contract** (drop unused) after dual-read window.

Service migration folders under each `services/*/migrations` or `db/migrations` — apply via approved migrator in CI job / platform-ops, never ad-hoc.

## Validation

| Check | Gate |
|-------|------|
| Schema version == expected | fail deploy |
| Replica lag < 5s | fail cutover |
| Row counts smoke on critical tables | alert |
| Constraint / FK integrity sample | fail if broken |
| Seed / reference data checksums (demo/training only) | env-specific |

## Rollback

- PITR to bookmark if expand migration is catastrophic.
- Prefer forward-fix compensating migration.
- Never `DROP` in emergency without CISO.

## Replication / DR

- Prod: sync replica + cross-region replica (async).
- DR overlay uses promote-replica runbook (`ops/runbooks/db-failover.md`).
- Backup CronJob: `infra/k8s/base/backup-cronjob.yaml` (alert `BackupMissed`).
