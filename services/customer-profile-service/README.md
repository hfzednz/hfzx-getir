# NEXORA Customer Profile Service

Go microservice for customer profile, preferences, addresses, consents, household, CRM notes/timeline, segments, personalization, AI model scores, and privacy workflows.

Module: `github.com/nexora/customer-profile-service`

**Hard rule:** Authentication, credentials, sessions, and IAM RBAC live only in `identity-service`. This service stores profile & CRM data keyed by `principal_id` and trusts gateway headers (`X-Nexora-User`, `X-Nexora-Tenant` / `X-Tenant-Id`).

## Quick start (dev / in-memory)

```bash
cd services/customer-profile-service
go run ./cmd/customer-profile-service
```

With empty `DATABASE_URL` the process boots in **dev mode** using in-memory repositories.

Health:

```bash
curl http://localhost:8081/health
curl http://localhost:8081/ready
```

Provision + me:

```bash
curl -s -X POST http://localhost:8081/v1/profile/customers \
  -H 'Content-Type: application/json' \
  -d '{"tenantId":"11111111-1111-1111-1111-111111111111","principalId":"22222222-2222-2222-2222-222222222222","displayName":"Ada"}'

curl -s http://localhost:8081/v1/profile/me \
  -H 'X-Nexora-User: 22222222-2222-2222-2222-222222222222' \
  -H 'X-Nexora-Tenant: 11111111-1111-1111-1111-111111111111'
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8081` | Listen address |
| `GRPC_ADDR` | `:9091` | Reserved for gRPC |
| `DATABASE_URL` | empty | Postgres DSN; empty → memory repos |
| `REDIS_URL` | empty | Redis URL; empty → in-process cache |
| `KAFKA_BROKERS` | empty | Comma-separated brokers (stub publisher) |
| `SEARCH_URL` | empty | OpenSearch URL (`OPENSEARCH_URL` alias); empty → indexer no-op |
| `CORS_ALLOWED_ORIGINS` | `*` | CORS |
| `RATE_LIMIT_PER_MINUTE` | `120` | Per-IP rate limit |
| `MEDIA_SERVICE_URL` | empty | Optional media-service base |
| `GEOFENCE_URL` | empty | Optional geofence-service base |

See `configs/config.example.yaml`.

## Docker Compose

```bash
docker compose up --build -d postgres redis customer-profile-service
# optional:
docker compose --profile kafka up -d redpanda
docker compose --profile search up -d opensearch
```

Migrations in `migrations/` are mounted into Postgres init.

## API surface

REST base: `/v1/profile/...` — profiles, me, addresses, preferences, avatar, tags, household, consents, CRM 360/notes/timeline, segments, personalization, AI model, privacy, admin search/merge/duplicates, health/ready.

- OpenAPI: `api/openapi/profile-v1.yaml`
- gRPC proto: `proto/profile/v1/profile.proto` (`GetProfile`, `GetPreferences`, `ListAddresses`, `CheckConsent`, `GetPersonalization`)

Error envelope (camelCase):

```json
{ "error": { "code": "not_found", "message": "...", "traceId": "...", "retriable": false } }
```

## Makefile

```bash
make test
make run
make build
make migrate   # requires DATABASE_URL + psql
```

## Layout

```
cmd/customer-profile-service/   main
internal/config/                env config
internal/domain/                entities + events
internal/app/                   use cases + ports + memory
internal/adapters/http|grpc|postgres|redis|kafka|search/
internal/observability/         metrics stubs
migrations/ api/openapi/ proto/ configs/
```
