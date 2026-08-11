# NEXORA SLO catalog

| Service | SLI | SLO | Window | Error budget |
|---------|-----|-----|--------|--------------|
| Edge / API gateway | successful requests | 99.95% | 30d | 0.05% |
| Checkout path | successful checkouts | 99.9% | 30d | 0.1% |
| Payment authorize | success (excl. user decline) | 99.9% | 30d | 0.1% |
| Dispatch assign | assign latency p99 < 2s | 99% | 30d | — |
| platform-ops-service | availability | 99.9% | 30d | 0.1% |
| liveops-service | flag eval latency p99 < 20ms | 99.9% | 30d | 0.1% |
| liveops-service | config resolve availability | 99.95% | 30d | 0.05% |
| supplier-service | onboarding/PO API availability | 99.9% | 30d | 0.1% |
| quality-service | gate evaluate / cert API availability | 99.9% | 30d | 0.1% |
| global-service | resolve latency p99 < 50ms | 99.9% | 30d | 0.1% |
| open-platform-service | webhook delivery success | 99.5% | 30d | 0.5% |
| superapp-service | shell resolve availability | 99.9% | 30d | 0.1% |
| innovation-service | simulation API availability | 99.5% | 30d | 0.5% |
| enterprise-ops-service | executive dashboard availability | 99.9% | 30d | 0.1% |
| hyperscale-cert-service | certification API availability | 99.9% | 30d | 0.1% |
| autonomy-service | genesis / heal API availability | 99.9% | 30d | 0.1% |
| Edge (GA regions) | successful requests | 99.99% | 30d | 0.01% |
| identity-service | auth API availability | 99.99% | 30d | 0.01% |
| order-service | place success (excl. client errors) | 99.95% | 30d | 0.05% |

Burn alerts: 2% budget in 1h → page; 5% in 6h → ticket.

Production alert rules: `infra/observability/prometheus/rules/nexora-production.yml`.
