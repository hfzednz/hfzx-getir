# NEXORA Order Service

Canonical OMS for NEXORA: order aggregate, place-order saga, compensations, returns/refunds, admin ops, and transactional outbox.

**Out of scope:** stock ledger (`inventory-service`), pick/pack tasks (`warehouse-service`), PSP charges (`payment-service`), courier matching (`dispatch-service`).

## Quick start (in-memory dev mode)

```bash
cd services/order-service
go test ./...
go run ./cmd/order-service
# HTTP :8086 — no DATABASE_URL required
```

```bash
curl -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  http://localhost:8086/v1/orders/health
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8086` | REST listen address |
| `GRPC_ADDR` | `:9096` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Stub cache |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop stub when empty) |
| `OPENSEARCH_URL` | *(empty)* | OpenSearch cluster (in-process fallback) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/orders/...`

- **Create / get / list / search** — draft or checkout create (idempotent)
- **Place** — saga: validate → soft reserve → authorize → confirm hard → start fulfillment
- **Cancel** — policy + Release + Void/Refund compensations
- **Returns / refunds** — request flows (payment refund port)
- **Lifecycle events** — warehouse / dispatch event application
- **Admin** — priority, split, guarded intervene, timeline
- **Outbox** — `POST /v1/orders/outbox/publish` drain stub
- **Health** — `/health` `/ready`

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

Tenant-scoped routes require `X-Tenant-Id`. Mutation idempotency via `Idempotency-Key` header or body field.

Money is **integer minor units** + ISO-4217 currency only (never float).

OpenAPI: `api/openapi/order-v1.yaml`  
gRPC proto: `proto/order/v1/order.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/order-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema (001–012)
```

## Docker

```bash
make docker-up
make docker-down
```
