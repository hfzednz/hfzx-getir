# Super App Runbook

## Health

- `GET /health` / `GET /ready`
- Admin: `GET /v1/superapp/admin/stats` with `X-Tenant-Id`

## Bootstrap

1. `POST /v1/superapp/bootstrap/mini-apps` — seed 17 mini-app catalog
2. `POST /v1/superapp/installs` — `{subjectId, moduleKey}`
3. `GET /v1/superapp/shell/resolve?subjectId=&shellVersion=`

## Hot update / rollback

1. Publish new manifest: `POST /v1/superapp/modules/{id}/manifests`
2. `POST /v1/superapp/installs/update`
3. On incident: `POST /v1/superapp/installs/rollback`

## Permissions

Grant only allow-listed permissions (`payments`, `navigation`, `search`, …). Unknown → sandbox violation.

## Outbox

`POST /v1/superapp/outbox/publish` drains pending `superapp.events`.

## Escalation

Does not own LiveOps flags or open-platform API keys — escalate to those services for enable/auth issues.
