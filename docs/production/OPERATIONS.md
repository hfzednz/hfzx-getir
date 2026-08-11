# Operations

## On-call

- Schedule: `ops/production/oncall.md`
- Escalation: `ops/production/escalation.md`
- Primary → secondary → SRE lead → CTO (SEV-1)

## Incident management

1. Detect (alert / customer / store review spike)
2. ACK + war room (`#nexora-incidents`)
3. Mitigate (flag, rollback, scale)
4. Communicate (status page)
5. Resolve + postmortem ≤ 5 business days for SEV-1/2

Playbooks: `ops/playbooks/`  
Runbooks: `ops/runbooks/`

## Maintenance windows

- `ops/production/maintenance-windows.md`
- Prefer expand migrations; announce ≥ 72h for customer-visible downtime (target: zero).

## Capacity planning

- `ops/production/capacity-planning.md`
- Inputs: orders/min forecast, city launch calendar, k6 load-env results.
