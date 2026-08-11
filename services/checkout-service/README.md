# NEXORA Checkout Service

Checkout session orchestration for NEXORA: start from cart, patch delivery prefs, run the validation pipeline, preview quotes, complete via `order-service` CreateFromCheckout, and recover abandoned sessions.

**Out of scope:** PSP capture (`payment-service`), OMS place saga (`order-service` owns it), inventory ledger (`inventory-service`), product master (`catalog-service`).

## Quick start (in-memory dev mode)

```bash
cd services/checkout-service
go test ./...
go run ./cmd/checkout-service
# HTTP :8088 — no DATABASE_URL required
```

Demo cart seeded in memory:

| Field | Value |
|-------|-------|
| Tenant | `11111111-1111-1111-1111-111111111111` |
| User | `22222222-2222-2222-2222-222222222222` |
| Cart | `33333333-3333-3333-3333-333333333333` |

```bash
curl -s -X POST http://localhost:8088/v1/checkout/sessions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -H "X-Nexora-User: 22222222-2222-2222-2222-222222222222" \
  -H "Idempotency-Key: demo-1" \
  -d '{"cartId":"33333333-3333-3333-3333-333333333333","deliveryOption":"instant"}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8088` | REST listen address |
| `GRPC_ADDR` | `:9098` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Stub cache |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop stub when empty) |
| `MIN_ORDER_MINOR` | `0` | Fallback minimum order (minor units) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/checkout/...`

- **POST /sessions** — start from cart (idempotent)
- **PATCH /sessions/{id}** — address / slot / gift / invoice / notes / substitutions / tip
- **POST /sessions/{id}/validate** — ordered validation pipeline
- **POST /sessions/{id}/refresh-quote** — PricingClient preview
- **POST /sessions/{id}/complete** — idempotent CreateFromCheckout
- **POST /sessions/{id}/abandon** — mark abandoned + recovery token
- **POST /recover** — recover abandoned session
- **GET /admin/metrics** — abandonment funnel
- **POST /outbox/publish** — drain stub
- **Health** — `/health` `/ready`

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

Tenant-scoped routes require `X-Tenant-Id`. Shopper identity via `X-Nexora-User` (body `principalId` fallback for tests). Mutation idempotency via `Idempotency-Key` header or body field.

Money is **integer minor units** + ISO-4217 currency only (never float).

OpenAPI: `api/openapi/checkout-v1.yaml`  
gRPC proto: `proto/checkout/v1/checkout.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/checkout-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema (001–003)
```

## Docker

```bash
make docker-up
make docker-down
```
