# Capacity Planning Guide

| Scenario | Peak RPS | Notes |
|----------|----------|-------|
| Black Friday | 50,000 | 3× normal; pre-warm HPA |
| Holiday | 35,000 | 48h sustained |
| Marketing spike | 25,000 | 15m burst |
| Regional spike | 20,000 | city launch |
| DR failover | 15,000 | absorption in secondary region |

Seed via `POST /v1/hyperscale/bootstrap` → capacity list.
