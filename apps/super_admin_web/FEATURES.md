# NEXORA Super Admin — Feature status

Status legend: **Complete** = route + feature module + mock BFF fallback; **Partial** = scaffold only; **Planned** = nav only.

| Module | Route | Status | Notes |
|--------|-------|--------|-------|
| Dashboard | `/dashboard` | Complete | Global KPIs, charts, alerts |
| Tenants | `/tenants`, `/tenants/[id]` | Complete | Isolation modes; dual-control suspend/delete |
| Companies | `/companies`, `/companies/[id]` | Complete | Legal entity settings |
| Countries | `/countries`, `/countries/[id]` | Complete | Locale / currency / timezone |
| Organization | `/org` | Complete | Users, departments, teams |
| Roles | `/roles` | Complete | Global role templates |
| Config | `/config` | Complete | Platform + brand settings |
| Feature flags | `/flags` | Complete | Rollouts; kill switches dual-control |
| Licenses | `/licenses` | Complete | Plans, limits, enterprise |
| Billing | `/billing` | Complete | FinOps / tenant billing |
| Security | `/security` | Complete | Threats, sessions, SSO/2FA policies |
| Compliance | `/compliance` | Complete | GDPR / KVKK / CCPA |
| Infrastructure | `/infra` | Complete | Clusters, K8s, DNS, CDN, certs |
| Databases | `/databases` | Complete | PG / Redis / OS / CH, replication |
| API gateway | `/gateway` | Complete | Keys, OAuth, rate limits |
| Messaging | `/messaging` | Complete | Kafka / queues / DLQ |
| Observability | `/observability` | Complete | Logs/metrics/traces, SLO/SLA |
| AI platform | `/ai-platform` | Complete | Models, inference, guardrails |
| Analytics | `/analytics` | Complete | Worldwide KPI aggregates only |
| Disaster recovery | `/disaster-recovery` | Complete | Backups, geo-repl, dual-control failover, restore, tests, simulations |
| Deployments | `/deployments` | Complete | CI/CD, blue-green/canary/rolling, rollback, secrets metadata (no values) |
| Monitoring | `/monitoring` | Complete | CPU/mem/disk/DB/API/queues/WS/Redis/OpenSearch/K8s/cloud |
| Notifications | `/notifications` | Complete | Provider hub + templates + delivery (not city inbox) |
| Audit | `/audit` | Complete | Immutable who/when/where/device/old/new/IP/session |
| Reports | `/reports` | Complete | Platform/infra/fin/compliance/security/AI/tenant + CSV/JSON/PDF mock export |

## Explicitly out of scope

| Concern | Where it belongs |
|---------|------------------|
| City orders / live map | `admin_web` |
| Courier dispatch | `admin_web` |
| Warehouse pick boards | `admin_web` |
| CRM / support ticket inbox | `admin_web` |
| City notification inbox | `admin_web` |

## Dual-control coverage

| Action | Module |
|--------|--------|
| `kill_switch` | Flags |
| `tenant_suspend` / `tenant_delete` | Tenants |
| `dr_failover` | Disaster recovery |
| `secret_rotate` | Deployments |
| `license_override` | Licenses |
