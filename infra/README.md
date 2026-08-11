# Infrastructure

Local deps:

```bash
docker compose -f infra/docker/docker-compose.yml up -d
```

Terraform (module contracts):

```bash
cd infra/terraform/envs/staging
terraform init
terraform plan
```

Helm:

```bash
helm lint infra/helm/nexora
helm upgrade --install nexora infra/helm/nexora -n nexora-apps --create-namespace
```

GitOps: `infra/argocd/applicationset.yaml`

Control plane API: `services/platform-ops-service` (`:8110`)

See `ARCHITECTURE.md`.
