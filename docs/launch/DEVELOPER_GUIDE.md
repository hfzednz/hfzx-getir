# Developer Guide — Integration

1. Read constitution `docs/constitution/MASTER_BLUEPRINT.md`
2. Service ports: `docs/launch/service-registry.yaml`
3. Local deps: `infra/docker/docker-compose.yml`
4. Run a service: `cd services/<name> && make test && make run`
5. Edge journeys: start `bff-customer` `:8111` (memory stubs by default)
6. Cert: `go run ./tools/integration-cert`
7. Errors: `docs/api/error-codes.md`
8. Money: int64 minor units; opaque cross-service IDs
