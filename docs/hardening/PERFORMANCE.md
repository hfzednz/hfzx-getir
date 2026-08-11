# Performance Guide

## Budgets (p99)

| Path | Target |
|------|--------|
| Checkout API | ≤150ms |
| Search | ≤80ms |
| Payment authorize | ≤200ms |
| Tracking update fanout | ≤100ms |
| BFF aggregate | ≤250ms |

## Throughput targets

| Metric | Target |
|--------|--------|
| Orders/sec | ≥5,000 |
| Search/sec | ≥20,000 |
| Payments/sec | ≥3,000 |
| AI/sec | ≥500 |
| Notifications/sec | ≥10,000 |
| Delivery updates/sec | ≥15,000 |
| DB TPS | ≥25,000 |

## Cache hierarchy

1. Edge CDN (static)  
2. Redis (session/hot keys)  
3. App in-process (short TTL)  
4. DB read replicas  

Configs: `infra/hardening/*` · Benches: `qa/hyperscale/`
