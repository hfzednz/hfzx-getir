# NEXORA Catalog Service

Enterprise PIM / headless product catalog for NEXORA. Owns master products, variants, SKUs, taxonomy, localized content, media references, publish workflow, and search indexing — **not** inventory quantities, sell prices, or orders.

## Quick start (in-memory dev mode)

```bash
cd services/catalog-service
go test ./...
go run ./cmd/catalog-service
# HTTP :8082 — no DATABASE_URL required
```

```bash
curl -H "X-Tenant-Id: 11111111-1111-1111-1111-111111111111" \
  http://localhost:8082/v1/catalog/health
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8082` | REST listen address |
| `GRPC_ADDR` | `:9092` | gRPC stub listen address |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; empty enables in-memory repos |
| `REDIS_URL` | *(empty)* | Query cache (stub) |
| `KAFKA_BROKERS` | *(empty)* | Comma-separated brokers (logs-only stub when empty) |
| `OPENSEARCH_URL` | *(empty)* | OpenSearch cluster (in-process fallback stub) |
| `MEDIA_SERVICE_URL` | `http://localhost:8085` | media-service base URL for CDN resolution |

See `configs/config.example.yaml` for a YAML-oriented view.

## API

REST base path: `/v1/catalog/...`

- **Products** — CRUD, archive, workflow (submit → approve → publish → hide)
- **Variants / SKUs** — option axes, barcode lookup (`GET /skus/lookup?type=ean&value=...`)
- **Categories, brands, attributes** — taxonomy & schema-driven values
- **Locales & SEO** — localized content and metadata
- **Media** — attach/detach refs (CDN via media-service port)
- **Bundles & relations** — composition and merchandising links
- **Versions** — list, diff, rollback snapshots on publish
- **Import** — CSV validate (`POST /import/validate`)
- **Search** — query, suggest, reindex (OpenSearch port)
- **Admin** — explorer, bulk status, duplicate detection
- **AI ports** — describe, translate, categorize, quality (stubs)

Errors use the NEXORA envelope: `{ "error": { "code", "message", "traceId", "retriable" } }`.

OpenAPI: `api/openapi/catalog-v1.yaml`  
gRPC proto: `proto/catalog/v1/catalog.proto`

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [FEATURES.md](./FEATURES.md).

```
cmd/catalog-service/main.go
internal/{domain,app,adapters,config}
migrations/          # PostgreSQL schema (001–018)
```

## Docker

```bash
make docker-up      # postgres + redis + catalog-service (memory mode)
make docker-down
```

With Kafka / OpenSearch profiles:

```bash
docker compose --profile kafka --profile search up --build -d
```

## Migrations

```bash
export DATABASE_URL=postgres://nexora:nexora@localhost:5433/catalog?sslmode=disable
make migrate
```

## Module

`github.com/nexora/catalog-service` · Go 1.22+
