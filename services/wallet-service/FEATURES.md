# Wallet Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| GetOrCreate | ✅ | Creates cash/refund/promo/cashback/gift accounts |
| Credit / Debit | ✅ | Idempotent; debit checks available |
| Hold / Release | ✅ | Hold blocks debit of reserved amount |
| Transfer | ✅ | Same or cross-wallet (same tenant) |
| History | ✅ | Paginated entries |
| Admin adjust | ✅ | Audited credit/debit with reason |
| Available = balance − held | ✅ | Never negative available |
| LedgerClient stub | ✅ | PostJournal on credit/debit |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8090` `/v1/wallets/...` | ✅ | NEXORA error envelope |
| Outbox PublishPending | ✅ | Stub drain |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC / Kafka | 🔶 | Proto + stub publisher |
| Loyalty points engine | ❌ | Opaque promo/cashback balances only |

## Explicit non-goals

- Order/cart/inventory engines
- Loyalty points accrual engine (balances may be credited by other services)
- Finance double-entry ownership (`finance-ledger-service` — stub port)

## Kafka topic

| Topic | Events |
|-------|--------|
| `wallet.lifecycle` | WalletCredited, WalletDebited, WalletHeld, WalletReleased, WalletTransferred, WalletAdjusted |
