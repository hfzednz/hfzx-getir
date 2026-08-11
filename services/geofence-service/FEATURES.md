# Geofence Service — Features

| Feature | Status |
|---------|--------|
| Create / update / delete zone | ✅ |
| Polygon (ray casting) contains | ✅ |
| Radius contains (haversine) | ✅ |
| Serviceability (city + point) | ✅ |
| Restricted zone blocks | ✅ |
| Zone list / get | ✅ |
| Outbox + ZoneChanged | ✅ |
| HTTP `/v1/geofence` + X-Tenant-Id | ✅ |
| Memory repos (dev) | ✅ |
| Postgres / Kafka / gRPC stubs | 🔶 |

## Kafka topics

- `geofence.zone` — `ZoneChanged`

## Non-goals

- Order / warehouse / inventory ownership
- Courier mobile UI
- Live tracking enter/exit (tracking-service)
