# NEXORA Production Deployment — Index (Prompt-43)

Architecture frozen. This pack prepares and operates global production release.

| Document | Purpose |
|----------|---------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Deployment topology by environment |
| [ENVIRONMENTS.md](./ENVIRONMENTS.md) | Prod / staging / QA / demo / DR / sandbox / load / training / regional |
| [RELEASE_PLAN.md](./RELEASE_PLAN.md) | SemVer, changelog, promotion path |
| [ROLLOUT.md](./ROLLOUT.md) | Blue/green, canary, feature, regional, global |
| [ROLLBACK.md](./ROLLBACK.md) | Image, migration, flag, store rollback |
| [CHECKLIST.md](./CHECKLIST.md) | Pre/during/post production checklist |
| [DATABASE_RELEASE.md](./DATABASE_RELEASE.md) | Migration order, PITR, verification |
| [VALIDATION.md](./VALIDATION.md) | Health gates for API/data/providers |
| [MONITORING.md](./MONITORING.md) | Prometheus/Grafana/Loki/Tempo + KPIs |
| [ALERTING.md](./ALERTING.md) | Alert catalog + routing |
| [AUTOSCALING.md](./AUTOSCALING.md) | HPA / Cluster Autoscaler / KEDA / GPU |
| [DISASTER_RECOVERY.md](./DISASTER_RECOVERY.md) | Regional/global failover |
| [SECURITY_VALIDATION.md](./SECURITY_VALIDATION.md) | TLS, secrets, WAF, Zero Trust |
| [COMPLIANCE.md](./COMPLIANCE.md) | GDPR / KVKK / PCI-DSS / store |
| [OPERATIONS.md](./OPERATIONS.md) | On-call, incidents, maintenance |
| [AI_OPS.md](./AI_OPS.md) | Model/prompt deploy + drift |
| [BUSINESS_VALIDATION.md](./BUSINESS_VALIDATION.md) | Journey smoke matrix |
| [POST_RELEASE.md](./POST_RELEASE.md) | Crash/ANR/revenue watch |
| [CERTIFICATE.md](./CERTIFICATE.md) | Production release certificate |
| [mobile/](./mobile/) | Android / iOS / ASO |

**GitOps SoT:** `infra/argocd`, `infra/k8s/overlays/*`, `infra/helm/nexora`  
**CI/CD:** `.github/workflows/cd-gitops.yml`, `cd-mobile.yml`, `cd-production-validate.yml`  
**Ops SoT:** `ops/release`, `ops/runbooks`, `ops/production`, `ops/playbooks`
