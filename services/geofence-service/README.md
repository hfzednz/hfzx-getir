# Geofence Service

Delivery / restricted / warehouse zone polygons and radius checks. Opaque ids only — no order, warehouse, or courier UI ownership.

## Quick start

```bash
go test ./...
go build ./...
HTTP_ADDR=:8099 go run ./cmd/geofence-service
```

## Examples

```bash
curl -s -X POST http://localhost:8099/v1/geofence/zones \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"name":"kadikoy","city":"Istanbul","kind":"delivery","vertices":[{"lat":-1,"lng":-1},{"lat":-1,"lng":1},{"lat":1,"lng":1},{"lat":1,"lng":-1}]}'

curl -s -X POST http://localhost:8099/v1/geofence/serviceability \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"city":"Istanbul","point":{"lat":0,"lng":0}}'
```

## Out of scope

Order aggregates, warehouse pick/pack, inventory ledgers, courier mobile UI.
