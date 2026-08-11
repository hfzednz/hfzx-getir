# Tracking Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| IngestLocation | ✅ | Latest + capped history + LocationUpdated |
| GetLiveCourier | ✅ | Latest GPS |
| GetOrderTimeline / AppendTimeline | ✅ | Opaque order_id |
| DetectArrival | ✅ | Haversine ≤ threshold → Arrived timeline |
| ListNearby | ✅ | Radius filter over latest locations |
| GeofenceClient | ✅ | Stub; records enter/exit when hits returned |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8098` `/v1/tracking/...` | ✅ | NEXORA error envelope; X-Tenant-Id |
| Outbox PublishPending | ✅ | Location / geofence / timeline topics |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Redis latest cache | 🔶 | Config stub |
| Kafka | 🔶 | Stub publisher |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- Order aggregate (`order-service`) — opaque `order_id` only
- Warehouse pick/pack (`warehouse-service`)
- Inventory / stock ledger
- Courier mobile UI (`apps/mobile_courier`)
- Zone polygon ownership (`geofence-service`)

## Kafka topics

| Topic | Events |
|-------|--------|
| `tracking.location` | LocationUpdated |
| `tracking.geofence` | GeofenceEnter, GeofenceExit |
| `tracking.timeline` | Arrived |
