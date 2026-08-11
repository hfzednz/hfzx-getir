# NEXORA Tracking Service

Live courier GPS, capped location history, delivery timeline projections, and arrival detection by distance threshold.

**Out of scope:** order aggregate (`order-service`), pick/pack (`warehouse-service`), stock ledger, courier mobile UI. Opaque refs only: `order_id`, `courier_id`.

**Arrival:** distance to dropoff ≤ threshold (default 50 m) or geofence enter (via GeofenceClient port).

## Quick start (in-memory dev mode)

```bash
cd services/tracking-service
go test ./...
go run ./cmd/tracking-service
# HTTP :8098 — no DATABASE_URL required
```

```bash
# Ingest GPS
curl -s -X POST http://localhost:8098/v1/tracking/locations \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"courierId":"cccccccc-cccc-cccc-cccc-cccccccccccc","lat":41.01,"lon":28.98,"accuracyM":10}'

# Live location
curl -s http://localhost:8098/v1/tracking/couriers/cccccccc-cccc-cccc-cccc-cccccccccccc/live \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8098` | REST listen address |
| `GRPC_ADDR` | `:9108` | gRPC stub address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `HISTORY_CAP` | `100` | Max history rows per courier |
| `ARRIVAL_THRESHOLD_M` | `50` | Arrival distance threshold |
| `GEOFENCE_BASE_URL` | `http://localhost:8099` | geofence-service stub target |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/tracking/...` — requires `X-Tenant-Id`.

- **POST /locations** — ingest GPS (updates latest + capped history)
- **GET /couriers/{id}/live** — latest courier location
- **GET /orders/{orderId}/timeline** — delivery timeline
- **POST /orders/{orderId}/timeline** — append timeline event
- **POST /arrival/detect** — arrival by distance threshold
- **GET /nearby?lat=&lon=&radiusM=** — couriers near a point
- **Health** — `/health` `/ready`

OpenAPI: `api/openapi/tracking-v1.yaml`  
gRPC proto: `proto/tracking/v1/tracking.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
