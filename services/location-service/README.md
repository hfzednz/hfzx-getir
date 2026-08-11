# NEXORA Location Service

Central **location infrastructure**: geocoding, addresses, spatial POI index, provider facade, caches, heatmaps.

**Out of scope:** dispatch assignment (`dispatch-service`), order lifecycle, live courier GPS SoT (`tracking-service`), geofence polygons SoT (`geofence-service`), route persistence SoT (`routing-service`). Map tiles are **not** served — client SDKs only; this service is a provider facade.

**Composes** `GeofenceClient` + `RoutingClient` ports for zone serviceability and route/ETA proxies.

## Quick start (in-memory dev mode)

```bash
cd services/location-service
go test ./...
go run ./cmd/location-service
# HTTP :8100 — no DATABASE_URL required
```

```bash
# Forward geocode
curl -s -X POST http://localhost:8100/v1/location/geocode \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"query":"Taksim Square"}'

# Nearby POIs
curl -s -X POST http://localhost:8100/v1/location/spatial/nearby \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{"lat":41.0082,"lng":28.9784,"radiusM":1000,"limit":10}'
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8100` | REST listen address |
| `GRPC_ADDR` | `:9110` | gRPC stub address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Redis GEO for hot nearby (memory Haversine fallback) |
| `OPENSEARCH_URL` | *(empty)* | OpenSearch geo_point dual-write indexer |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (noop when empty) |
| `GEOFENCE_URL` | *(empty)* | geofence-service base (memory stub) |
| `ROUTING_URL` | *(empty)* | routing-service base (memory stub) |

See `configs/config.example.yaml`.

## API

REST base path: `/v1/location/...` — requires `X-Tenant-Id`. NEXORA error envelope.

| Area | Endpoints |
|------|-----------|
| Geocode | `POST /geocode`, `/geocode/reverse`, `/geocode/autocomplete` |
| Addresses | `POST /addresses/validate`, `/normalize`, `/enrich` |
| Spatial | `POST /pois`, `/spatial/nearby`, `/radius`, `/bbox`, `/nearest` |
| Routes | `POST /routes`, `/eta` → routing-service |
| Zones | `POST /zones/serviceability` → geofence-service |
| History | `POST/GET /history` (privacy-scoped, capped) |
| Maps | `GET/POST /maps/offline...` (manifests only) |
| AI / Admin | `GET /ai/heatmap`, `GET /admin/coverage` |

OpenAPI: `api/openapi/location-v1.yaml`  
gRPC proto: `proto/location/v1/location.proto`

**Privacy:** handlers never log full precise coordinates at info level.

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

## Docker

```bash
docker compose up --build -d
```
