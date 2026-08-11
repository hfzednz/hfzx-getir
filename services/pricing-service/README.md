# NEXORA Pricing Service

Price books (base / regional / warehouse / customer / vip / corporate), tax display rules, dynamic adjustments, and quote assembly via **PromoClient.Evaluate**.

**Out of scope:** catalog ownership, inventory ledger, promotions storage (`promotion-service`), payments / loyalty. Money is **int64 minor units**. `variant_id` is an opaque UUID.

## Quick start (in-memory dev mode)

```bash
cd services/pricing-service
go test ./...
go run ./cmd/pricing-service
# HTTP :8095 — no DATABASE_URL required
```

```bash
# Create price book
curl -s -X POST http://localhost:8095/v1/pricing/books \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"name":"default","currency":"TRY"}'

# Upsert base price, then quote
curl -s -X POST http://localhost:8095/v1/pricing/quote \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"currency":"TRY","lines":[{"variantId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","qty":1}]}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8095` | REST listen address |
| `GRPC_ADDR` | `:9105` | gRPC stub address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Redis stub |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop when empty) |
| `PROMO_BASE_URL` | `http://localhost:8094` | promotion-service Evaluate stub target |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/pricing/...` — requires `X-Tenant-Id`, optional `X-Nexora-User`.

- **Books / Prices** — upsert book, upsert scoped entry, resolve waterfall
- **Quote / Simulate** — assemble lines + dynamic + promo + tax + fees
- **Tax / Dynamic** — calculate tax, apply dynamic, upsert rules
- **Admin** — list books, entries, rules, audits
- **Health** — `/health` `/ready`

OpenAPI: `api/openapi/pricing-v1.yaml`  
gRPC proto: `proto/pricing/v1/pricing.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
