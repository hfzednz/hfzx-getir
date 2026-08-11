# Autonomy dependency graph snapshot

Living graph is owned by `autonomy-service` (`GET /v1/autonomy/dependencies` after bootstrap).

## Core edges (representative)

```text
bff-customer → order-service (http)
bff-customer → catalog-service (http)
bff-customer → payment-service (http)
order-service → inventory-service (http)
order-service → payment-service (http)
dispatch-service → tracking-service (kafka)
liveops-service → bff-customer (port)
superapp-service → liveops-service (port)
innovation-service → liveops-service (port)
enterprise-ops-service → security-service (port)
hyperscale-cert-service → quality-service (port)
hyperscale-cert-service → platform-ops-service (port)
autonomy-service → hyperscale-cert-service (port)
autonomy-service → platform-ops-service (port)
autonomy-service → quality-service (port)
autonomy-service → security-service (port)
autonomy-service → liveops-service (port)
```

Full registry: `docs/launch/service-registry.yaml`.
