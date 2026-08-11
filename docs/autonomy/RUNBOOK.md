# Autonomy runbook

## Bootstrap + Genesis

```bash
export TENANT=<uuid>
curl -X POST -H "X-Tenant-Id: $TENANT" http://localhost:8125/v1/autonomy/bootstrap
curl -H "X-Tenant-Id: $TENANT" http://localhost:8125/v1/autonomy/gates
curl -X POST -H "X-Tenant-Id: $TENANT" -H "Content-Type: application/json" \
  -d '{"version":"1.0.0"}' http://localhost:8125/v1/autonomy/genesis
```

## Self-heal

```bash
curl -X POST -H "X-Tenant-Id: $TENANT" -H "Content-Type: application/json" \
  -d '{"kind":"service","targetRef":"bff-customer","action":"restart"}' \
  http://localhost:8125/v1/autonomy/heal
```

Actual K8s mutation is executed via **platform-ops-service** port — autonomy records the plan/result only.

## Incident correlation

1. Observe alerts (Prometheus / platform-ops)
2. Autonomy heal action with matching `targetRef`
3. AI CTO review for recurrence (`POST /v1/autonomy/reviews`)
4. Evolution task if systemic (`POST /v1/autonomy/evolution`)

## Rollback

Release meta plans support `canary` / `blue_green`; apply/rollback remains GitOps/platform-ops.
