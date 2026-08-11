# Deployment Architecture

## Principles

- GitOps only (Argo CD ApplicationSet → Kustomize overlays → Helm umbrella where referenced).
- No manual kubectl apply to prod except break-glass (logged via platform-ops).
- Edge BFFs + Envoy/Istio ingress; domain services private.
- Data plane: PostgreSQL (primary + replica), Redis, Kafka, OpenSearch, ClickHouse (analytics), object storage.
- Observability plane: Prometheus, Grafana, Loki, Tempo, Alertmanager (`nexora-obs`).

## Topology (logical)

```text
                    [ CDN / WAF ]
                          |
                   [ Ingress / Mesh ]
                    /      |      \
            bff-*     realtime    open-platform
               |          |            |
         domain services (identity…autonomy)
               |          |            |
         Postgres / Redis / Kafka / Search / AI
```

## Environments → overlays

| Environment | Overlay | Cluster intent |
|-------------|---------|----------------|
| development | `infra/k8s/overlays/dev` | Shared or local compose |
| QA | `infra/k8s/overlays/qa` | Integration + quality gates |
| staging | `infra/k8s/overlays/staging` | Prod-like soak |
| demo | `infra/k8s/overlays/demo` | Sales / partner demos |
| sandbox | `infra/k8s/overlays/sandbox` | Partner API experiments |
| load | `infra/k8s/overlays/load` | k6 / chaos / capacity |
| training | `infra/k8s/overlays/training` | Ops training tenants |
| production | `infra/k8s/overlays/prod` | Live traffic |
| disaster recovery | `infra/k8s/overlays/dr` | Warm standby region |

## Regional model

- Active-active edge + active-passive data for DR region (see DISASTER_RECOVERY.md).
- Country/city residency via `global-service` policies; DB schemas remain opaque tenant IDs.

## Control planes (do not redesign)

| Plane | Owner |
|-------|--------|
| Infra apply / scale / backup | `platform-ops-service` `:8110` |
| Flags / kill switches | `liveops-service` `:8116` |
| Quality cert | `quality-service` `:8118` |
| Hyperscale cert | `hyperscale-cert-service` `:8124` |
| Autonomy / genesis | `autonomy-service` `:8125` |
| Enterprise BCP | `enterprise-ops-service` `:8123` |

## Image registry

`ghcr.io/nexora/<service>:<semver|sha|env-tag>` promoted only via `cd-gitops` PR + Argo sync.
