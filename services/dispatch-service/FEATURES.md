# Dispatch Service — Features

| Feature | Status |
|---------|--------|
| Create dispatch (queued) | ✅ |
| Auto-assign nearest + capacity | ✅ |
| Manual assign / reassign | ✅ |
| Pickup → transit → arrive | ✅ |
| Complete with POD (otp/qr/photo/signature/gps) | ✅ |
| Fail + optional requeue | ✅ |
| Batch create | ✅ |
| Fleet vehicle upsert / list | ✅ |
| Courier availability snapshot | ✅ |
| Admin list | ✅ |
| Illegal transition guard | ✅ |
| Outbox + domain events | ✅ |
| Routing / tracking / geofence stubs | ✅ |
| HTTP `/v1/dispatch` + X-Tenant-Id | ✅ |
| Memory repos (dev) | ✅ |
| Postgres / Kafka / gRPC stubs | 🔶 |

## Kafka topics

- `dispatch.job` — DispatchCreated, Pickup*, Delivery*
- `dispatch.assignment` — CourierAssigned, CourierReassigned

## Non-goals

- Order / warehouse / inventory ownership
- Courier mobile UI (`apps/mobile_courier`)
- Real routing engine / live GPS stream (routing / tracking services)
