# Monitoring

## Stack

| Component | Path / location |
|-----------|-----------------|
| Prometheus | `infra/observability/prometheus/` |
| Rules | `rules/nexora-slo.yml`, `rules/nexora-production.yml` |
| Grafana dashboards | `infra/observability/grafana/dashboards/` |
| Tempo | `infra/observability/tempo/tempo.yaml` |
| Loki | via observability Terraform module |
| Alertmanager | `infra/observability/alertmanager/alertmanager.yml` |

## Dashboards (required)

1. **Platform Overview** — `platform-overview.json`
2. **Release Health** — canary weight, 5xx, deploy events
3. **Business KPIs** — orders/min, GMV minor units, authorize rate, ETA breach
4. **Data plane** — Postgres connections, Redis memory, Kafka lag
5. **AI** — inference latency, drift score, GPU util

## SLI / SLO

Canonical: `ops/slo/catalog.md`  
Target platform availability: **99.99%** edge for GA regions (error budget burn alerts).

## Tracing

- OTLP → Tempo (`OTEL_EXPORTER_OTLP_ENDPOINT` in Helm env).
- Trace critical path: BFF → identity/catalog/cart/checkout/order/payment/dispatch.
