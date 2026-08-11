# Prompt-43 production pack index (ops)

- Docs: `docs/production/`
- Overlays: `infra/k8s/overlays/{dev,qa,staging,demo,sandbox,load,training,prod,dr}`
- Argo: `infra/argocd/applicationset.yaml`
- Helm prod values: `infra/helm/nexora/values/prod.yaml`
- Alerts: `infra/observability/prometheus/rules/nexora-production.yml`
- Alertmanager: `infra/observability/alertmanager/alertmanager.yml`
- Workflows: `cd-mobile`, `cd-production-validate`, `release-changelog`
- Mobile Fastlane: `apps/mobile_*/fastlane`
- ASO: `store/aso/`
- Validate: `tools/prod-validate`
