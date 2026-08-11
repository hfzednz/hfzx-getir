# Promotion Service

NEXORA `promotion-service` — campaigns, promotions, coupons, vouchers, rule engine, and cart evaluate.

## Run

```bash
make run          # HTTP :8094, in-memory repos
make test
make build
```

Module: `github.com/nexora/promotion-service`

## HTTP

Base path: `/v1/promo/...`  
Required header: `X-Tenant-Id` (UUID)  
Errors: NEXORA `{ "error": { "code", "message", "traceId", "retriable" } }`

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/promo/campaigns` | Create campaign |
| POST | `/v1/promo/campaigns/{id}/activate\|pause\|expire` | Lifecycle |
| POST | `/v1/promo/promotions` | Create promotion + rule |
| POST | `/v1/promo/coupons` | Generate coupon |
| POST | `/v1/promo/coupons/redeem` | Redeem (idempotent) |
| POST | `/v1/promo/vouchers` | Issue voucher |
| POST | `/v1/promo/vouchers/redeem` | Redeem (idempotent) |
| POST | `/v1/promo/evaluate` | Core cart evaluate |
| POST | `/v1/promo/simulate` | Dry-run + persist |
| GET | `/v1/promo/admin/overview` | Admin counts |

Money is always **int64 minor units** + ISO-4217 currency. Variant/category/brand/segment ids are opaque.

## Config

See `configs/config.example.yaml`. Empty `DATABASE_URL` → memory mode.

## Explicit non-goals

No PSP, loyalty points, order aggregate, inventory ledger, or catalog product master.
