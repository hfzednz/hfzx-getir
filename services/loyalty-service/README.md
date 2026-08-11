# NEXORA Loyalty Service

Points, memberships, rewards, referrals, cashback orchestration, and gamification (missions, streaks, achievements, spin).

**Out of scope:** PSP payments (`payment-service`), coupon/campaign engines (`promotion-service`), CRM tickets, wallet ledger ownership. Cashback/promo money credits go through **WalletClient.Credit**. Points/XP stay in the loyalty DB. Money amounts are **int64 minor units**. `principal_id` / `order_id` are opaque UUIDs.

## Quick start (in-memory dev mode)

```bash
cd services/loyalty-service
go test ./...
go run ./cmd/loyalty-service
# HTTP :8093 — no DATABASE_URL required
```

```bash
curl -s -X POST http://localhost:8093/v1/loyalty/accounts \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -H "X-Nexora-User: 22222222-2222-2222-2222-222222222222" \
  -d '{}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8093` | REST listen address |
| `GRPC_ADDR` | `:9103` | gRPC stub address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Redis stub |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop when empty) |
| `WALLET_BASE_URL` | `http://localhost:8090` | Wallet credit stub target |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/loyalty/...` — requires `X-Tenant-Id`, optional `X-Nexora-User`.

- **Accounts** — ensure/get, points history
- **Points** — earn (order idempotent), redeem
- **Membership** — get, evaluate upgrade
- **Rewards** — list, unlock, redeem
- **Cashback** — grant (pending), confirm → `WalletClient.Credit`
- **Referrals** — code, apply (blocks self), complete
- **Gamification** — missions, streaks, spin, achievements
- **AI / Leaderboard** — stubs
- **Admin** — manual grant (audited)
- **Health** — `/health` `/ready`

OpenAPI: `api/openapi/loyalty-v1.yaml`  
gRPC proto: `proto/loyalty/v1/loyalty.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
