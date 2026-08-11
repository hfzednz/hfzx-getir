# Platform Ops Service

HTTP `:8110` `/v1/platform` — deployments, scaling, backups, recovery, alerts, cost, SLO.

Does not own application domains. Pairs with `infra/` GitOps and Terraform.

```bash
make test && make run
```
