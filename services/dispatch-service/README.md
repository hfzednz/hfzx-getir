# Dispatch Service

Real-time DMS: create jobs, auto/manual assign, reassign, pickup→delivery lifecycle with POD, fleet vehicles, courier availability snapshots. Opaque `order_id` / `fulfillment_id` / `warehouse_id` / `courier_principal_id` only.

## Quick start

```bash
go test ./...
go build ./...
HTTP_ADDR=:8096 go run ./cmd/dispatch-service
```

## Examples

```bash
# Upsert courier snapshot
curl -s -X POST http://localhost:8096/v1/dispatch/couriers/snapshot \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"courierPrincipalId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","available":true,"lat":41.01,"lng":29.0,"currentLoad":0,"maxCapacity":3,"rating":4.9,"vehicleType":"bike","onShift":true}'

# Create dispatch
curl -s -X POST http://localhost:8096/v1/dispatch/jobs \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"orderId":"22222222-2222-2222-2222-222222222222","pickup":{"lat":41.011,"lng":29.001},"dropoff":{"lat":41.05,"lng":29.05},"requiredVehicle":"bike","city":"Istanbul"}'
```

## Out of scope

Order aggregates, warehouse pick/pack, inventory ledgers, courier mobile UI.
