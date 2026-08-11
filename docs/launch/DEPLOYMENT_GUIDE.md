# Deployment Guide

Canonical production pack: [`docs/production/README.md`](../production/README.md).

## Local

```bash
docker compose -f infra/docker/docker-compose.yml up -d
cd services/bff-customer && make run
```

## Staging / Prod (GitOps)

1. Build & push service images (CI)
2. Update overlay tags (`.github/workflows/cd-gitops.yml`)
3. Argo CD sync `nexora-staging` / `nexora-prod` (ApplicationSet covers all envs)
4. Canary per `infra/k8s/overlays/prod/canary-analysis.yaml`
5. Validate: `go run ./tools/integration-cert` and `go run ./tools/prod-validate -env=staging|prod`
6. Sign GO: `ops/release/GO_NO_GO.md`

## Mobile

- Android AAB / iOS IPA: `.github/workflows/cd-mobile.yml`
- Store copy: `store/aso/`
- Guides: `docs/production/mobile/`

## Environments

dev · qa · staging · demo · sandbox · load · training · prod · dr — see `docs/production/ENVIRONMENTS.md`.

Ports: `docs/launch/service-registry.yaml`.
