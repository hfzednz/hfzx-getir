# Warehouse Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| ReceiveFulfillment | ✅ | Projection + SoftReserve + pick task queued |
| ClaimPick / StartPick / Scan / CompletePick | ✅ | Barcode validate; remaining qty blocks complete |
| ClaimPack / VerifyWeight / Seal / GenerateLabel | ✅ | Creates dispatch unit |
| DispatchVerify / HandoffConfirm | ✅ | Consume via InventoryClient |
| Reassign / Cancel / Escalate tasks | ✅ | Append-only history |
| QC pass/fail | ✅ | Inspection lifecycle |
| ClockIn/Out + breaks | ✅ | Active shift tracking |
| Equipment heartbeat | ✅ | Online status |
| AI OptimizeRoute | ✅ | Location-sort stub |
| Dashboard aggregates | ✅ | Counts by status |
| Kafka events | ✅ | Log/noop stub without brokers |
| InventoryClient SoftReserve/Confirm/Release/Consume | ✅ | Port only — no ledger |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| gRPC serve | 🔶 | Proto defined; listener stub |
| Redis task queues | 🔶 | Stub client |
| Stock ledger / order SoT / catalog | ❌ | Out of scope by design |

## Explicit non-goals

- On-hand / reserved qty ledger (`inventory-service`) — only InventoryClient port
- Canonical order state machine (`order-service`)
- Product master / prices (`catalog-service` / `pricing-service`)
- Courier matching (`dispatch-service`) — warehouse confirms handoff only

## Fulfillment pipeline

```
Received → Reserved (SoftReserve) → PickQueued → Picking → Picked
→ PackQueued → Packing → Packed → DispatchQueued → Dispatched (Consume)
```

## Kafka topics

| Topic | Events |
|-------|--------|
| `warehouse.fulfillment.events` | OrderReceived, PickingStarted/Completed, PackingStarted/Completed, LabelGenerated, DispatchStarted/Completed, CourierAssigned, QCPassed/Failed |
| `warehouse.task.events` | TaskAssigned, TaskCompleted, TaskEscalated, TaskCancelled, TaskReassigned |
