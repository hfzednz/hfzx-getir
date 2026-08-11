# NEXORA Payment Service

PCI-DSS-ready payment orchestration: intents, authorize/capture/void, PSP routing with failover, refunds, chargebacks, fraud scoring port, and eligibility for checkout (no charge).

**Out of scope:** order/cart/inventory/loyalty engines. Opaque `order_id` refs only. Money is **int64 minor units** + ISO-4217.

## Quick start (in-memory dev mode)

```bash
cd services/payment-service
go test ./...
go run ./cmd/payment-service
# HTTP :8089 — no DATABASE_URL required
```

```bash
curl -s -X POST http://localhost:8089/v1/payments/intents \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -H "X-Nexora-User: 22222222-2222-2222-2222-222222222222" \
  -H "Idempotency-Key: demo-pay-1" \
  -d '{"orderId":"ord-opaque-1","amountMinor":5000,"currency":"TRY","methodType":"card"}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8089` | REST listen address |
| `GRPC_ADDR` | `:9099` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop stub when empty) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/payments/...`

- **POST /intents** — create intent (idempotent)
- **GET /intents/{id}** — get intent
- **GET /intents** — admin list
- **POST /intents/{id}/authorize** — fraud → PSP/wallet authorize
- **POST /intents/{id}/capture** — full or partial capture
- **POST /intents/{id}/void** — release authorization
- **POST /intents/{id}/refund** — partial/full refund
- **POST /eligibility** — methods available (no charge)
- **GET|POST /routes** — provider routing
- **POST /intents/{id}/chargebacks** — record dispute
- **POST /outbox/publish** — drain stub
- **Health** — `/health` `/ready`

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

OpenAPI: `api/openapi/payment-v1.yaml`  
gRPC proto: `proto/payment/v1/payment.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/payment-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema (001–003)
```

## Docker

```bash
docker compose up --build -d
```
