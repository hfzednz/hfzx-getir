# Disaster Recovery

## Objectives

| Metric | Target |
|--------|--------|
| RPO | ≤ 5 minutes (async replica); critical ledger ≤ 1 minute where sync available |
| RTO | ≤ 30 minutes regional; ≤ 2 hours global DNS+store status |

## Regional failover

1. Detect primary region SEV-1 (multi-AZ failure).
2. platform-ops initiates failover playbook.
3. Promote Postgres replica in DR (`ops/runbooks/db-failover.md`).
4. Point Kafka MirrorMaker / consumers to DR (or activate standby cluster).
5. Switch DNS / Anycast to DR ingress.
6. Validate VALIDATION.md probes.
7. Communicate status page + store messaging if customer-impacting.

## Global failover

- Traffic manager health checks on `/ready` of BFFs.
- Read-only mode flag (LiveOps) if payments/PSP region dark.

## Recovery testing

- Quarterly DR drill on `infra/k8s/overlays/dr` — signed in GO_NO_GO.
- Monthly backup restore to sandbox.
- Chaos CronJob: `infra/k8s/base/chaos-cronjob.yaml` (non-prod first).

## Data plane recovery

| System | Method |
|--------|--------|
| PostgreSQL | PITR / replica promote |
| Redis | rebuild from primary + warm cache |
| Kafka | mirrored topics + consumer reset policy |
| Object storage | cross-region replication |
| OpenSearch | snapshot restore |
