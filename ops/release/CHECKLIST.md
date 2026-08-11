# Release Checklist

1. Tag RC: `vYYYY.MM.DD-rcN`
2. Run `go run ./tools/integration-cert` from repo root
3. Path-filtered CI green (services + infra)
4. Promote images via `.github/workflows/cd-gitops.yml` → staging
5. Canary 10% → 25% → 50% → 100% (see `infra/k8s/overlays/prod/canary-analysis.yaml`)
6. Verify Grafana Platform Overview + alert silence hygiene
7. Smoke journeys: customer order, courier offer, warehouse pick/pack, admin dashboard
8. Freeze feature flags except kill switches
9. Publish release notes + rollback GitOps tag documented
10. On GO: sync prod Application; watch error budget 2h
