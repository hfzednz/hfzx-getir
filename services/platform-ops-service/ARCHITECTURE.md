# NEXORA Platform Ops Service

SRE / platform control plane — deployments, scaling, backups, recovery, alerts, cost snapshots, SLO burn.

- HTTP `:8110` `/v1/platform`
- Does **not** own app business logic; records and orchestrates infrastructure intents via ports (K8s/Argo/cloud).

See `ARCHITECTURE.md` under `infra/` for full cloud design.
