# settlement-service

NEXORA courier/supplier/merchant/partner settlement batches, dual-control approval, payouts, and reconciliation.

- **Module:** `github.com/nexora/settlement-service`
- **HTTP:** `:8092` (`/v1/settlements/...`)
- **Money:** `int64` minor units + ISO-4217 (never float)
- **Hard rules:** no cart/order/inventory; opaque payee/business refs only

## Quick start

```bash
go test ./...
go build ./...
HTTP_ADDR=:8092 go run ./cmd/settlement-service
```

Leave `DATABASE_URL` empty for in-memory repositories (dev/tests).

## Lifecycle

`draft` → `pending_approval` → `approved` → `paying` → `completed` | `failed`

Dual-control: approver must differ from submitter (`ErrDualControl`).

## Key endpoints

| Method | Path | Use case |
|--------|------|----------|
| POST | `/v1/settlements/batches` | CreateBatch |
| POST | `/v1/settlements/batches/{id}/lines` | AddLine |
| POST | `/v1/settlements/batches/{id}/submit` | Submit |
| POST | `/v1/settlements/batches/{id}/approve` | Approve (dual-control) |
| POST | `/v1/settlements/batches/{id}/execute` | ExecutePayouts |
| POST | `/v1/settlements/batches/{id}/reconcile` | ReconcileProviderReport |
| GET | `/v1/settlements/batches` | List |

Headers: `X-Tenant-Id` (required), `X-Nexora-User` (actor), `Idempotency-Key` on create.

## Ports

- `LedgerClient` — posts settlement journals to finance-ledger-service
- `PayoutClient` — bank/PSP stub
- `EventPublisher` — outbox → `settlement.lifecycle`

## Docs

- `ARCHITECTURE.md` — FinTech platform binding
- `FEATURES.md` — feature matrix
- `api/openapi/settlement-v1.yaml`
- `proto/settlement/v1/settlement.proto`
