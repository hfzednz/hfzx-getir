# Payment Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| CreateIntent | ✅ | Idempotent by key; opaque order_id |
| Authorize | ✅ | Fraud port → MockPSP / Failover / wallet debit |
| Capture / PartialCapture | ✅ | Remaining capturable tracking |
| Void | ✅ | Authorized only, uncaptured |
| Refund partial/full | ✅ | Idempotent; remaining refundable |
| Eligibility | ✅ | No charge; method list for checkout |
| RouteProvider / UpsertRoute | ✅ | Ordered PSP preference |
| RecordChargeback | ✅ | Opens dispute + event |
| Admin list | ✅ | Filter by status/principal/order |
| MockPSP + Failover | ✅ | Primary fail → secondary |
| Fraud block | ✅ | Decision block / high score |
| LedgerClient stub | ✅ | PostJournal on money paths |
| WalletClient pay-with-wallet | ✅ | Debit port |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8089` `/v1/payments/...` | ✅ | NEXORA error envelope |
| Outbox PublishPending | ✅ | Stub drain |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Kafka | 🔶 | Stub publisher |
| Order/cart/inventory engines | ❌ | Out of scope by design |

## Explicit non-goals

- Order aggregate / OMS saga (`order-service`)
- Cart / inventory / loyalty points engines
- Raw card PAN storage (tokens only)
- Finance double-entry ownership (`finance-ledger-service` — stub port only)

## State machine

```
initiated → authorized | failed
authorized → captured | voided | failed
captured → refunded (full); partial refund keeps captured
```

## Kafka topic

| Topic | Events |
|-------|--------|
| `payment.lifecycle` | PaymentInitiated, PaymentAuthorized, PaymentCaptured, PaymentFailed, PaymentVoided, RefundRequested, RefundCompleted, ChargebackCreated |
