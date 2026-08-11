# finance-ledger-service

NEXORA double-entry finance ledger — chart of accounts, immutable journals, tax, invoices.

- **Module:** `github.com/nexora/finance-ledger-service`
- **HTTP:** `:8091` (`/v1/ledger/...`)
- **Money:** `int64` minor units + ISO-4217 (never float)
- **Hard rules:** no cart/order/inventory aggregates; opaque external refs only

## Quick start

```bash
go test ./...
go build ./...
HTTP_ADDR=:8091 go run ./cmd/finance-ledger-service
```

Leave `DATABASE_URL` empty for in-memory repositories (dev/tests).

## Key endpoints

| Method | Path | Use case |
|--------|------|----------|
| POST | `/v1/ledger/accounts` | EnsureAccount |
| GET | `/v1/ledger/accounts/{id}/balance` | GetBalance |
| POST | `/v1/ledger/journals` | PostJournal (must balance) |
| GET | `/v1/ledger/journals` | ListJournals |
| POST | `/v1/ledger/invoices` | CreateInvoice |
| POST | `/v1/ledger/invoices/{id}/credit-notes` | IssueCreditNote |
| POST | `/v1/ledger/tax/calculate` | TaxCalculate |
| POST | `/v1/ledger/outbox/publish` | Drain outbox |

Headers: `X-Tenant-Id` (required), `Idempotency-Key` on money mutations.

## Domain invariants

- Journal posts require ≥2 lines with `sum(debit) == sum(credit)`.
- Posted journals are immutable.
- Invoice statuses: `draft` → `issued` → `paid` \| `void` \| `credited`.

## Docs

- `ARCHITECTURE.md` — FinTech platform binding
- `FEATURES.md` — feature matrix
- `api/openapi/ledger-v1.yaml`
- `proto/ledger/v1/ledger.proto`
