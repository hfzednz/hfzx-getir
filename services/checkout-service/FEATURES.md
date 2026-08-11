# Checkout Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| StartFromCart | ✅ | Idempotent by key; loads cart via CartClient |
| Patch prefs | ✅ | address / slot / gift / invoice / notes / substitutions / tip |
| Validate pipeline | ✅ | customer → zone → ATP → price → coupon → restrictions → fraud → pay elig → duplicate → min order |
| RefreshQuote | ✅ | PricingClient preview (minor units) |
| Complete | ✅ | Idempotent → OrderClient.CreateFromCheckout (opaque order_id) |
| Optional PlaceOrder | ✅ | Port only; OMS owns place saga |
| RecoverAbandoned | ✅ | Recovery token rotate |
| Admin list / metrics | ✅ | Status funnel + abandon rate |
| Memory repos + clients | ✅ | Injectable zone/inventory failures for tests |
| Concurrent complete lock | ✅ | Per-key CompleteLock |
| HTTP `:8088` `/v1/checkout/...` | ✅ | NEXORA error envelope, X-Tenant-Id, X-Nexora-User |
| Outbox PublishPending | ✅ | Stub drain via EventPublisher |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Kafka / Redis | 🔶 | Stub adapters |
| PSP capture / OMS saga / stock ledger | ❌ | Out of scope by design |

## Explicit non-goals

- Payment capture (`payment-service`) — eligibility port only
- Place-order saga ownership (`order-service`) — CreateFromCheckout (+ optional PlaceOrder port)
- Inventory ledger mutations (`inventory-service`) — ATP check only
- Product master (`catalog-service`)

## Validation pipeline (ordered)

```
customer → address/zone → inventory ATP → price refresh → coupon
→ age/region → fraud/risk → payment eligibility → duplicate → min order
```

## State machine

```
started → validating → ready | blocked
ready → completing → completed | failed
blocked → validating (refresh)
* → abandoned (recover → started)
```

## Kafka topic

| Topic | Events |
|-------|--------|
| `checkout.lifecycle` | CheckoutStarted, CheckoutValidated, CheckoutCompleted, CheckoutFailed, CheckoutAbandoned, CheckoutRecovered |
