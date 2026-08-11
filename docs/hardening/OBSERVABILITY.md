# Observability Optimization Guide

- Prefer exemplars linking traces↔metrics for p99 paths
- Structured JSON logs; sample debug at ≤1%
- Alert on SLO burn (existing `ops/slo/catalog.md`) not raw CPU alone
- Hyperscale dashboard: cert gates + benchmark trend
- Cardinality limits on custom tags
