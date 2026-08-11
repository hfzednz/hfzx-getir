# Pricing Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| UpsertPriceBook / UpsertPrice | ✅ | Scopes: base, regional, warehouse, customer, vip, corporate |
| GetPrice waterfall | ✅ | base→regional→warehouse→customer→vip→corporate |
| QuoteCart assemble | ✅ | Lines + fees + tax + promo discounts |
| ApplyDynamic | ✅ | percent/fixed; time_of_day or inventory_hint |
| TaxCalculate | ✅ | Rate in basis points; display only |
| SimulateQuote | ✅ | Simulated flag; no QuoteCreated event |
| Admin list | ✅ | Books, entries, tax, dynamic, audits |
| PromoClient.Evaluate | ✅ | Port + stub/mock; no local promo storage |
| DynamicHintClient | ✅ | Optional stub |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8095` `/v1/pricing/...` | ✅ | NEXORA error envelope; X-Tenant-Id |
| Outbox PublishPending | ✅ | Stub drain |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Kafka / Redis / Promo HTTP | 🔶 | Stub publishers/clients |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- Catalog / product master (`catalog-service`) — opaque `variant_id` only
- Inventory ledger (`inventory-service`) — hints via DynamicHintClient only
- Promotions storage / coupons (`promotion-service`) — Evaluate only
- Payments / PSP (`payment-service`)
- Loyalty points (`loyalty-service`)

## Kafka topics

| Topic | Events |
|-------|--------|
| `pricing.price` | PriceChanged |
| `pricing.quote` | QuoteCreated |
