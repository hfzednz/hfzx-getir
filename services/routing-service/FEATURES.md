# Routing Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| CreateRoute | ✅ | Waypoints + initial Haversine ETA |
| Optimize (nearest-neighbor) | ✅ | Origin fixed; stops reordered |
| RecalculateETA | ✅ | Optional courier move; traffic/weather factors |
| UpdateTrafficHint | ✅ | Regional radius factor |
| GetRoute | ✅ | By tenant + id |
| ETA = dist/speed × traffic × weather | ✅ | Haversine; DefaultSpeed ~30 km/h |
| MapsClient | ✅ | Haversine distance-matrix stub |
| TrafficClient | ✅ | Fixed-factor stub |
| WeatherClient | ✅ | HTTP when `WEATHER_URL`/`OPENWEATHER_URL` set; else factor 1.0 |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8097` `/v1/routing/...` | ✅ | NEXORA error envelope; X-Tenant-Id |
| Outbox PublishPending | ✅ | RouteUpdated / ETAUpdated |
| PostgreSQL adapters | 🔶 | Migrations ready; repos stub |
| Kafka | 🔶 | Stub publisher |
| gRPC | 🔶 | Proto + stub server |

## Explicit non-goals

- Order aggregate (`order-service`) — opaque `order_id` only
- Warehouse pick/pack (`warehouse-service`)
- Inventory / stock ledger
- Courier mobile UI (`apps/mobile_courier`)
- Payments / tips settlement

## Kafka topics

| Topic | Events |
|-------|--------|
| `routing.route` | RouteUpdated |
| `routing.eta` | ETAUpdated |
