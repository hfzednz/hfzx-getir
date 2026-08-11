# NEXORA cart-service

Real-time shopping cart persistence for guest and authenticated shoppers: lines, coupons, pricing quotes, soft reservations, merge-on-login, abandoned-cart recovery, and recommendations.

**Port:** `:8087` · **Base path:** `/v1/cart/...` · **Module:** `github.com/nexora/cart-service`

## Boundaries

| Owns | Does not own |
|------|----------------|
| Cart aggregate, lines, applied coupons, quote snapshot, soft-reserve refs, saved carts, wishlist links | Product master (`catalog-service`), stock ledger (`inventory-service`), orders (`order-service`), PSP charges (`payment-service`) |

Money is always **int64 minor units** + ISO-4217 currency. `variant_id` is opaque.

## Quick start

```bash
make run          # HTTP_ADDR=:8087, in-memory mode
make test
make build
```

```bash
# Create guest cart
curl -s -X POST http://localhost:8087/v1/cart \
  -H 'X-Tenant-Id: 11111111-1111-1111-1111-111111111111' \
  -H 'X-Guest-Token: guest-abc' \
  -H 'Content-Type: application/json' \
  -d '{"currency":"TRY"}'

# Add line
curl -s -X POST http://localhost:8087/v1/cart/{cartId}/lines \
  -H 'X-Tenant-Id: 11111111-1111-1111-1111-111111111111' \
  -H 'Content-Type: application/json' \
  -d '{"variantId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","qty":2}'
```

## Headers

| Header | Required | Purpose |
|--------|----------|---------|
| `X-Tenant-Id` | yes | Tenant UUID |
| `X-Guest-Token` | guest flows | Anonymous shopper token |
| `X-Nexora-User` | auth flows | Principal UUID |
| `X-Request-Id` | optional | Trace id |
| `Idempotency-Key` | soft-reserve | Idempotent inventory hold |

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

## Environment

| Variable | Default | Notes |
|----------|---------|-------|
| `HTTP_ADDR` | `:8087` | HTTP listen |
| `GRPC_ADDR` | `:9097` | gRPC stub |
| `DATABASE_URL` | empty | Empty = in-memory repos |
| `REDIS_URL` | empty | Stub |
| `KAFKA_BROKERS` | empty | Stub publisher → `cart.lifecycle` |
| `PRICING_URL` | empty | Stub PricingClient |
| `INVENTORY_URL` | empty | Stub InventoryClient |
| `CORS_ALLOWED_ORIGINS` | `*` | CORS |
| `RATE_LIMIT_PER_MINUTE` | `240` | Per-IP |

## Layout

```text
cmd/cart-service/          main
internal/domain/           cart, line, quote, coupon, merge, events
internal/app/              use cases + ports + memory
internal/adapters/         http, kafka, postgres, redis, pricing, inventory, grpc
migrations/                001–010
api/openapi/cart-v1.yaml
proto/cart/v1/cart.proto
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).
