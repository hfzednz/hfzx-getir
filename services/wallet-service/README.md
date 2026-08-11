# NEXORA Wallet Service

Customer wallets with per-principal accounts (`cash`, `refund`, `promo`, `cashback`, `gift`), holds, transfers, and audited admin adjustments.

**Out of scope:** order/cart/inventory/loyalty engines. Money is **int64 minor units** + ISO-4217. `Available = balance − held` (never negative).

## Quick start (in-memory dev mode)

```bash
cd services/wallet-service
go test ./...
go run ./cmd/wallet-service
# HTTP :8090 — no DATABASE_URL required
```

```bash
curl -s -X POST http://localhost:8090/v1/wallets \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -H "X-Nexora-User: 22222222-2222-2222-2222-222222222222" \
  -d '{"currency":"TRY"}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8090` | REST listen address |
| `GRPC_ADDR` | `:9100` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop stub when empty) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/wallets/...`

- **POST /** — get or create wallet
- **GET /{id}** — wallet + accounts
- **POST /{id}/credit** — credit (idempotent)
- **POST /{id}/debit** — debit available
- **POST /{id}/hold** — reserve funds
- **POST /holds/{holdId}/release** — release hold
- **POST /{id}/transfer** — transfer between accounts/wallets
- **GET /{id}/history** — entry history
- **POST /{id}/admin/adjust** — audited adjust
- **POST /outbox/publish** — drain stub
- **Health** — `/health` `/ready`

OpenAPI: `api/openapi/wallet-v1.yaml`  
gRPC proto: `proto/wallet/v1/wallet.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
