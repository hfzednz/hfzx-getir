# Rollout Strategies

## Canary (default for critical path)

Contract: `infra/k8s/overlays/prod/canary-analysis.yaml`

| Step | Weight | Pause | Abort if |
|------|--------|-------|----------|
| 1 | 10% | 5m | error rate > 1% or success < 99.5% |
| 2 | 25% | 10m | SLO burn |
| 3 | 50% | 15m | payment/checkout regression |
| 4 | 100% | — | — |

Critical services (identity, order, payment, checkout, BFF customer): **canary mandatory**.  
Non-critical (innovation, demo tools): RollingUpdate OK.

## Blue / Green

- Used for gateway / Envoy config and irreversible data-plane cutovers.
- Green cluster/revision receives synthetic + employee traffic → DNS/weight flip → blue retained 24h for rollback.

## Feature rollout (LiveOps)

- Percentage / city / warehouse / segment / app version flags via `liveops-service`.
- Dual-control for kill switches (`bff-admin`).
- Never ship unpaid risky features without flag default OFF in prod.

## Regional rollout

1. Soft-launch city (flag + capacity verified)
2. Metro expansion
3. Country GA
4. Multi-country (global-service residency checks)

## Global rollout

1. Primary region prod canary 100%
2. Secondary region warm (DR overlay synced)
3. Cross-region read replicas verified
4. CDN/geo DNS live
5. Store phased release (Play 10% → 50% → 100%; App Store phased 7-day)
