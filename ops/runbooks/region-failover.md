# Region failover / DR

## RTO / RPO targets
- RTO: 30 minutes critical path
- RPO: 5 minutes (async cross-region backups)

## Steps
1. Confirm primary region health (cluster API, ingress, DNS)
2. Trigger DNS failover (Cloudflare / Route53 health checks)
3. Promote standby Postgres / Kafka consumers in secondary
4. Scale apps in secondary via Argo Application sync
5. Record `RecoveryStarted` in platform-ops-service
6. Run smoke: identity login, catalog browse, cart checkout dry-run

## Post-incident
- Update postmortem, restore primary when stable, reverse DNS
