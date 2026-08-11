# NEXORA Enterprise Operations Service

Org hierarchy, governance, PMO, strategy/OKRs, BCP, risk, audit, executive dashboards.

- HTTP `:8123` `/v1/enterprise`
- Kafka `enterprise.events`
- Does **not** own ERP GL, security GRC SoT, analytics warehouse, or platform-ops infra

```bash
make test && make run
```
