# Rollback Procedures

## Application image (preferred)

1. Revert GitOps overlay `newTag` to last known good SHA/tag (or revert merge of promote PR).
2. Argo CD sync / self-heal applies previous ReplicaSet.
3. Confirm `/ready` and error rate recovery within 5 minutes.
4. Post incident note in `ops/release/notes/`.

## Canary abort

- Automatic on `nexora-canary-policy` breach.
- Manual: set LiveOps kill switch for feature OR pin Rollout weight to 0 / previous stable.

## Database

- Prefer **expand/contract**; never destructive down-migration in prod without CISO+CTO.
- If migration fails mid-flight: stop deploy, restore from pre-migration snapshot / PITR to bookmark (DATABASE_RELEASE.md).
- Forward-fix with compensating migration when possible.

## Kafka / consumers

- Replay from consumer group offset bookmark taken pre-release.
- DLQ drain playbook: `ops/playbooks/kafka-dlq.md`.

## Mobile store

- Play: halt staged rollout; publish previous AAB if needed.
- App Store: pause phased release; ship hotfix build.
- Remote config / LiveOps can disable broken client paths without store wait.

## Break-glass

- platform-ops emergency scale / rollback APIs audited.
- All break-glass actions require dual-control + postmortem within 48h.
