# Location Service — Feature Matrix

| Area | Status | Notes |
|------|--------|-------|
| ForwardGeocode / Reverse / Autocomplete | ✅ | Cache on second call; MockMapsProvider |
| ValidateAddress / Normalize / Enrich | ✅ | Feasibility via GeofenceClient |
| UpsertPOI / Nearby / Nearest / Radius / BBox | ✅ | Haversine in memory |
| IngestHistory / GetHistory | ✅ | Capped at 100 per subject |
| Offline manifests | ✅ | Metadata only; no tiles |
| UpsertHeatCell / DemandHeatmap | ✅ | Grid demand/density |
| ProxyRoute / ProxyETA | ✅ | RoutingClient port |
| CheckZoneServiceability | ✅ | GeofenceClient port |
| AdminCoverageStats | ✅ | POI + heat aggregates |
| Memory repos | ✅ | Dev + tests |
| HTTP `:8100` `/v1/location/...` | ✅ | NEXORA errors; X-Tenant-Id |
| Outbox PublishPending | ✅ | location.address / location.geo / … |
| Maps / Geofence / Routing adapters | ✅ | Stubs (memory-backed) |
| PostgreSQL adapters | ✅ | Postgres repos when `DATABASE_URL` set |
| Redis GEO / OpenSearch | ✅ | Redis GEO when `REDIS_URL`; OpenSearch geo_point when `OPENSEARCH_URL` |
| Kafka | 🔶 | Stub publisher |
| gRPC | 🔶 | Proto + stub server |
| PostGIS GEOGRAPHY GiST | 🔶 | Documented in 009_indexes.sql |

## Explicit non-goals

- Dispatch assignment (`dispatch-service`)
- Order lifecycle (`order-service`)
- Live courier GPS store (`tracking-service`)
- Geofence polygon SoT (`geofence-service`) — composed via port
- Route persistence SoT (`routing-service`) — composed via port
- Map tile rendering / serving

## Kafka topics

| Topic | Events |
|-------|--------|
| `location.address` | AddressValidated, AddressCreated |
| `location.geo` | LocationUpdated, RouteCalculated, ETAUpdated |
| `location.geofence` | GeofenceEntered, GeofenceExited |
| `location.zone` | DeliveryZoneChanged |
