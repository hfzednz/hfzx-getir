# Order Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| CreateDraft / CreateFromCheckout | ✅ | Idempotent by key |
| PlaceOrder saga | ✅ | validate → soft reserve → authorize → confirm hard → fulfillment |
| Saga step retry stub | ✅ | One extra attempt per step |
| Payment fail compensation | ✅ | Release + payment_failed → cancelled |
| ApplyWarehouseEvent | ✅ | PickingStarted, PackingCompleted (+ RequestDispatch) |
| ApplyDispatchEvent | ✅ | CourierAssigned, OutForDelivery, Delivered→Completed |
| CancelOrder | ✅ | Policy + Release + Void |
| RequestReturn / RequestRefund | ✅ | Post-delivery; payment refund port |
| Admin SetPriority / Split / Intervene / Timeline | ✅ | Intervene still enforces transitions |
| Search orders | ✅ | Memory indexer |
| Outbox PublishPending | ✅ | Stub drain via EventPublisher |
| Memory clients (inv/pay/wh/dispatch) | ✅ | Injectable failures for tests |
| Concurrent place idempotency | ✅ | Per-key PlaceLock |
| HTTP `:8086` `/v1/orders/...` | ✅ | NEXORA error envelope, X-Tenant-Id |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Kafka / OpenSearch / Redis | 🔶 | Stub adapters |
| Live PSP / stock / pick tasks | ❌ | Out of scope by design |

## Explicit non-goals

- Stock ledger mutations (`inventory-service`) — SoftReserve/ConfirmHard/Release ports only
- Pick/pack task assignment (`warehouse-service`) — ReceiveFulfillment + event projection
- Card/wallet charging (`payment-service`) — Authorize/Void/Refund ports only
- Courier matching (`dispatch-service`) — RequestDispatch port only

## Place saga (happy path)

```
Validate → SoftReserve → AuthorizePayment → ConfirmHard → StartFulfillment
→ warehouse_assigned
→ (events) picking → packing → ready_for_dispatch → courier_assigned
→ out_for_delivery → delivered → completed
```

## Kafka topic

| Topic | Events |
|-------|--------|
| `order.lifecycle` | OrderCreated, OrderValidated, InventoryReserved, PaymentAuthorized, WarehouseAssigned, PickingStarted, PackingCompleted, CourierAssigned, OutForDelivery, Delivered, Completed, Cancelled, RefundRequested, RefundCompleted, … |
