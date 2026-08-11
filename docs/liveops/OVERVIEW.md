# LiveOps Platform — Prompt #31

## Service

| Item | Value |
|------|--------|
| Service | `liveops-service` |
| HTTP | `:8116` `/v1/liveops` |
| gRPC | `:9116` |
| Kafka | `liveops.events` |
| SoT | Flags, remote config, experiments, LiveOps calendar, rollouts/rollbacks |

## Non-ownership

- Campaigns/coupons → `promotion-service`
- Notification delivery → `notification-service`
- Analytics warehouse → `data-platform-service`
- AI model serving → `ai-platform-service` (winner-hint port only)

## Facades

`feature-flag-service` and `config-service` remain split-ready; clients should prefer `liveops-service`.

## Admin surfaces (API-backed)

LiveOps Dashboard · Feature Flag Console · Experiment Manager · Remote Config Manager · Rollout Dashboard · Approval Center · Growth Dashboard (`GET /v1/liveops/admin/stats` + domain APIs)
