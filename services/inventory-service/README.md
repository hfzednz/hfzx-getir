# NEXORA Inventory Service

Real-time inventory source of truth for NEXORA: warehouses, locations, stock balances, soft/hard reservations, ATP, FEFO lots, transfers, counts, returns, and forecast projections.

**Out of scope:** product titles (`catalog-service`), sell prices (`pricing-service`), order state machine (`order-service`).

## Quick start (in-memory dev mode)

```bash
cd services/inventory-service
go test ./...
go run ./cmd/inventory-service
# HTTP :8083 — no DATABASE_URL required
```

```bash
curl -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  http://localhost:8083/v1/inventory/health
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8083` | REST listen address |
| `GRPC_ADDR` | `:9093` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Soft-hold cache (stub) |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (logs-only stub when empty) |
| `OPENSEARCH_URL` | *(empty)* | OpenSearch cluster (in-process fallback stub) |
| `SOFT_RESERVE_TTL` | `15m` | Default soft reservation TTL |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/inventory/...`

- **Warehouses / locations** — CRUD location tree
- **Stock** — ensure, get/list, adjust/receive/damage/waste, movements ledger
- **Reservations** — soft, confirm-hard, extend, release, consume (idempotent)
- **ATP** — variant/warehouse/region availability query
- **Transfers** — create, approve, complete (moves stock)
- **Counts** — start, submit lines, approve variance
- **Lots** — near-expiry list, FEFO allocate
- **Returns** — receive with restock/quarantine/waste disposition
- **Forecast** — get/upsert + stub AI generate
- **Search** — index query / reindex
- **Health** — `/health` `/ready`

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

Tenant-scoped routes require `X-Tenant-Id`. Mutation idempotency via `Idempotency-Key` header or body field.

OpenAPI: `api/openapi/inventory-v1.yaml`  
gRPC proto: `proto/inventory/v1/inventory.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/inventory-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema (001–014)
```

## Docker

```bash
make docker-up      # postgres + redis + inventory-service (memory mode)
make docker-down
```

With Kafka / OpenSearch profiles:

```bash
docker compose --profile kafka --profile search up --build -d
```

## Migrations

```bash
export DATABASE_URL=postgres://nexora:nexora@localhost:5434/inventory?sslmode=disable
make migrate
```
