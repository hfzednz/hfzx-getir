# Inventory Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| Warehouses CRUD | ✅ | Soft-delete → closed |
| Locations tree | ✅ | Materialized path, zone types |
| EnsureBalance | ✅ | Zero-balance create |
| Adjust / Receive / Damage / Waste | ✅ | Idempotent movements |
| SoftReserve | ✅ | Per-key mutex, idempotent |
| ConfirmHard / Extend / Release / Consume | ✅ | Soft→Hard→Consumed state machine |
| Soft expire restores available | ✅ | On get/extend |
| Concurrent oversell guard | ✅ | Per stock-key mutex in memory |
| ATP query | ✅ | Policy: optional incoming/in-transit |
| Transfers create/approve/complete | ✅ | Moves on_hand between WH |
| Counts start/submit/approve | ✅ | Variance posts adjustments |
| Lots near-expiry + FEFO | ✅ | Used by soft reserve when `useFefo` |
| Returns receive | ✅ | restock / quarantine / waste |
| Forecast get/upsert/generate | ✅ | Stub AI port |
| Search index stock | ✅ | Memory + OpenSearch stub |
| Kafka events | ✅ | Log/noop stub without brokers |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Redis soft holds | 🔶 | Stub client |
| Catalog titles / prices / orders | ❌ | Out of scope by design |

## Explicit non-goals

- Product master content (`catalog-service`) — only opaque `variant_id` / `sku_code`
- Sellable prices or tax (`pricing-service`)
- Order aggregate / checkout state machine (`order-service`)
- Pick/pack task assignment (`warehouse-service`)

## Reservation state machine

```
Soft → Hard (confirm) → Consumed (ship deduct)
Soft → Released | Expired
Hard → Released
```

Invariant: cannot reserve more than `Available()` = `on_hand - reserved - blocked`.

## Kafka topics

| Topic | Events |
|-------|--------|
| `inventory.stock.events` | InventoryCreated/Updated, StockAdjusted, StockReceived, StockExpired |
| `inventory.reservation.events` | StockReserved, ReservationReleased/Expired/Confirmed/Consumed |
| `inventory.transfer.events` | StockTransferred |
| `inventory.count.events` | StockCountCompleted |
| `inventory.index.commands` | ReindexStock |
