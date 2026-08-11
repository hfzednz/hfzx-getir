# Alerting Catalog

## Routing

| Severity | Channel | Response |
|----------|---------|----------|
| critical | Pager (on-call) | 5m ACK |
| high | Slack `#nexora-incidents` | 15m |
| medium | Ticket | next business day |
| low | Dashboard only | — |

Config: `infra/observability/alertmanager/alertmanager.yml`

## Domain alerts (see Prometheus rules)

| Alert | Signal | Runbook |
|-------|--------|---------|
| HighErrorRate | 5xx > 1% 5m | `ops/runbooks/high-error-rate.md` |
| PaymentAuthorizeFail | authorize fail ratio | `ops/runbooks/payment-outage.md` |
| OrderPlaceLatency | p99 place > SLO | `ops/runbooks/high-error-rate.md` |
| KafkaConsumerLag | lag > threshold | `ops/runbooks/kafka-lag.md` |
| RedisMemoryHigh | used_memory > 85% | `ops/runbooks/redis-memory.md` |
| PostgresConnections | > 80% max | capacity playbook |
| AIInferenceLatency | p99 spike | `docs/production/AI_OPS.md` |
| NotificationDeliveryFail | channel fail | notification runbook |
| DeploymentFailed | deploy metric | rollback |
| BackupMissed | >48h | backup CronJob |
| MobileCrashSpike | Crashlytics rate | `ops/runbooks/mobile-crash-spike.md` |

Burn: 2% budget in 1h → page; 5% in 6h → ticket (`ops/slo/catalog.md`).
