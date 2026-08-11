# Enterprise Ops Runbook

1. Seed org: `POST /v1/enterprise/org`
2. Policies: upsert → `POST /v1/enterprise/policies/approve`
3. PMO: portfolio → program → project
4. Continuity: upsert → `POST /v1/enterprise/continuity/activate` on crisis
5. Executive: `GET /v1/enterprise/executive/dashboard?role=CEO`
6. Outbox: `POST /v1/enterprise/outbox/publish`

Security policy enforcement remains in `security-service` (port gate only).
