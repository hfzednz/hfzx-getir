# Runbook — Database failover

## Symptoms

- Primary Postgres unavailable / region loss

## Mitigate

1. Declare SEV-1; follow `docs/production/DISASTER_RECOVERY.md`.
2. Promote replica in DR; update connection secrets via ExternalSecrets.
3. Bounce dependent deployments (rollout restart) for pool refresh.
4. Switch DNS to DR ingress.
5. Validate `tools/prod-validate -env=dr`.
6. Keep primary as quarantined until forensics complete.
