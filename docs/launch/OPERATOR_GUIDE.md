# Operator Guide (Launch)

## Day-2 operations

- Dashboards: Grafana `nexora-platform` + service RED panels
- Alerts: `infra/observability/prometheus/rules/nexora-slo.yml`
- Deploy/rollback: Argo CD + `platform-ops-service` `/v1/platform/deployments`
- Secrets: Vault via `security-service` metadata (no plaintext in Git)
- Backups: Velero schedule + PG CronJob; verify weekly restore

## On-call

1. Page on Critical SLO burn / payment path
2. Runbook: `ops/runbooks/high-error-rate.md`
3. DR: `ops/runbooks/region-failover.md`
4. Record incident in security/platform-ops as appropriate

## Support handoff

- Customer tickets → CRM (`crm-service`) via `bff-customer` / admin
- Refunds: request-only through CRM ports (payment owns capture)
