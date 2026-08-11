# NEXORA Warehouse Service

Operations & fulfillment (WOMS): pick/pack/dispatch tasks, stations, workforce shifts, equipment heartbeats, QC, and AI route-optimize stubs.

**Out of scope:** stock ledger (`inventory-service` via InventoryClient port only), order aggregate (`order-service`), catalog content (`catalog-service`).

## Quick start (in-memory dev mode)

```bash
cd services/warehouse-service
go test ./...
go run ./cmd/warehouse-service
# HTTP :8085 — no DATABASE_URL required
```

```bash
curl -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  http://localhost:8085/v1/warehouse/health
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8085` | REST listen address |
| `GRPC_ADDR` | `:9095` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Task queue stub |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (logs-only stub when empty) |
| `INVENTORY_SERVICE_URL` | *(empty)* | inventory-service base URL (local stub when empty) |
| `WEIGHT_TOLERANCE_G` | `50` | Default pack weight tolerance |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/warehouse/...`

- **Fulfillment** — receive, get, list, cancel
- **Tasks** — queue list, claim pick, reassign, cancel, escalate
- **Picking** — start, scan line, complete → pack task
- **Packing** — claim station, verify weight, seal, generate label → dispatch unit
- **Dispatch** — queue, verify package, handoff confirm (inventory Consume)
- **QC** — create inspection, pass/fail
- **Workforce** — clock in/out, breaks, performance snapshot
- **Equipment / stations** — register, heartbeat, list
- **Sensors** — temperature/humidity ingest (stub)
- **AI** — route optimize stub
- **Admin** — dashboard aggregates
- **Health** — `/health` `/ready`

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

Tenant-scoped routes require `X-Tenant-Id`.

OpenAPI: `api/openapi/warehouse-v1.yaml`  
gRPC proto: `proto/warehouse/v1/warehouse.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/warehouse-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema stubs
```

## Docker

```bash
make docker-up
make docker-down
```

## Migrations

```bash
export DATABASE_URL=postgres://nexora:nexora@localhost:5435/warehouse?sslmode=disable
make migrate
```
