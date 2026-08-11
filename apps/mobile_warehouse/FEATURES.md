# Warehouse App — Feature Status

> Architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md) · Design density: **Dense**

## Modules

| Module | Status |
|--------|--------|
| Auth stub (phone → OTP → session) | Implemented (swap for full OTP later) |
| Home dashboard (`GET /warehouse/dashboard`) | Implemented |
| Picking queue/task/scan + `PickingRules` | Implemented |
| Packing weight/seal/label + `PackingRules` | Implemented |
| Dispatch handoff + QR + realtime invalidate | Implemented + `HandoffRules` |
| Inventory stock/adjust/cycle/inbound stub | Implemented + `InventoryRules` |
| Transfers create/list/approve | Implemented |
| Expiry near-list + waste | Implemented |
| Purchasing suppliers/PO receive+QC | Implemented |
| Returns customer/courier/supplier | Implemented |
| Quality QC queues | Implemented |
| Map layout viewer (CustomPainter) | Implemented |
| AI hub (forecast / path / restock) | Implemented |
| Shifts clock/break/attendance | Implemented |
| Tasks (cleaning/maintenance/emergency) | Implemented |
| Reports KPI cards | Implemented |
| Notifications / settings / support | Implemented |
| More hub | Implemented |
| Role soft-gates (`RoleRules`) | Implemented |

## Tests

`test/unit/business_rules/` — picking, packing, inventory, handoff, role.

## Routes

Wired in `lib/routing/app_router.dart` + `route_names.dart` (shell: home / picking / packing / dispatch / more).
