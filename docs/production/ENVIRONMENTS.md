# Environments

## Matrix

| Env | Purpose | Data | Traffic | Auto-sync Argo | Notes |
|-----|---------|------|---------|----------------|-------|
| development | Local / feature | ephemeral / DevMode OK | none | optional | `DATABASE_URL` empty allowed |
| QA | Integration + quality-service gates | synthetic | CI + QA | yes | Flaky quarantine enforced |
| staging | Soak + canary rehearsal | anonymized prod-like | internal | yes | ≥24h soak before prod GO |
| demo | Sales / investor | seeded demo tenants | restricted | yes | Resettable |
| sandbox | Partner OpenAPI | isolated | partner keys | yes | Rate-limited hard |
| load | Capacity / k6 / KEDA tune | synthetic high-volume | generators only | yes | No real PSP charges |
| training | Runbook drills | synthetic | training users | yes | Chaos optional |
| production | Live | real | public | yes (after GO) | Dual-control flags |
| DR | Failover target | replica / PITR | standby | yes | DNS cutover drill quarterly |

## Terraform

- Staging: `infra/terraform/envs/staging`
- Production: `infra/terraform/envs/prod` (node_min 6, node_max 80, GPU pool on)
- DR region: clone prod module with distinct CIDR (document in `DISASTER_RECOVERY.md`)

## Secrets

- Never in Git. Vault / cloud secret manager → ExternalSecrets → pods.
- Prod identity requires `JWT_KEY_PEM`, `OTP_PEPPER`, non-wildcard CORS (Prompt-42 gates).
- Payments require `STRIPE_SECRET_KEY` (or equivalent PSP); MockPSP forbidden in prod.

## Networking

- Prod ingress host: `api.nexora.example` (replace in Helm values-prod).
- mTLS mesh when `global.mesh.enabled=true`.
- NetworkPolicies in `infra/k8s/base/networkpolicy.yaml`.
