# Reliability Guide

## Patterns (adopt in BFFs/clients — do not rewrite domain services)

| Pattern | Guidance |
|---------|----------|
| Circuit breaker | Open after 5 consecutive 5xx; half-open after 30s |
| Retry | Exponential backoff, max 3, jitter; only retriable |
| Bulkhead | Separate pools for payment vs catalog vs search |
| Fallback | Cached catalog / degraded search on upstream fail |
| Graceful degradation | LiveOps flags for feature kill-switch |

Chaos metadata: `POST /v1/hyperscale/chaos` · packs in `qa/hyperscale/chaos/`
