# NEXORA Routing Service

Route plans, multi-stop optimize (nearest-neighbor), and ETA estimates with traffic/weather factors.

**Out of scope:** order aggregate (`order-service`), pick/pack (`warehouse-service`), stock ledger, courier mobile UI. Opaque refs only: `order_id`, `dispatch_id`, `courier_id`, `warehouse_id`.

**ETA formula:** `ETA = (distance / speed) * trafficFactor * weatherFactor` (Haversine distance; MapsClient stub).

## Quick start (in-memory dev mode)

```bash
cd services/routing-service
go test ./...
go run ./cmd/routing-service
# HTTP :8097 — no DATABASE_URL required
```

```bash
# Create multi-stop route
curl -s -X POST http://localhost:8097/v1/routing/routes \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"waypoints":[
    {"kind":"warehouse","lat":41.0082,"lon":28.9784,"label":"WH"},
    {"kind":"stop","lat":41.05,"lon":29.0,"label":"C"},
    {"kind":"stop","lat":41.015,"lon":28.985,"label":"A"},
    {"kind":"stop","lat":41.03,"lon":28.99,"label":"B"}
  ]}'

# Optimize then recalculate ETA after move
curl -s -X POST http://localhost:8097/v1/routing/routes/{id}/optimize \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8097` | REST listen address |
| `GRPC_ADDR` | `:9107` | gRPC stub address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop when empty) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/routing/...` — requires `X-Tenant-Id`.

- **POST /routes** — create draft route + initial ETA
- **GET /routes/{id}** — get route
- **POST /routes/{id}/optimize** — nearest-neighbor multi-stop
- **POST /routes/{id}/recalculate-eta** — refresh ETA (optional current lat/lon)
- **POST /traffic-hints** — upsert regional traffic factor
- **Health** — `/health` `/ready`

OpenAPI: `api/openapi/routing-v1.yaml`  
gRPC proto: `proto/routing/v1/routing.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
