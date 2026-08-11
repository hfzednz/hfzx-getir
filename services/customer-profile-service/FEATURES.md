# Customer Profile Service — Feature Status

| Feature | Status | Notes |
|---------|--------|-------|
| Domain model (profile, address, prefs, tags, household, consent, CRM, segments, personalization, AI, privacy, activity) | **Done** | `internal/domain` |
| App ports (hexagonal repositories + media/zone/events) | **Done** | `internal/app/ports` |
| App use cases on `Deps` | **Done** | provision, addresses, prefs, avatar, tags, household, consents, CRM 360, segments, personalization, AI recompute, privacy, admin merge/search |
| In-memory repos (dev / tests) | **Done** | `internal/app/memory` |
| Env config | **Done** | `HTTP_ADDR`, `DATABASE_URL`, `REDIS_URL`, `KAFKA_BROKERS`, `SEARCH_URL`/`OPENSEARCH_URL`, … |
| HTTP REST (ARCHITECTURE routes under `/v1/profile`) | **Done** | stdlib ServeMux Go 1.22+ → `Deps` |
| HTTP middleware (requestID, log, recover, CORS, rate limit) | **Done** | |
| Trusted principal (`X-Nexora-User` + tenant) for `/me` | **Done** | No IAM auth logic |
| NEXORA JSON error envelope | **Done** | camelCase `error.{code,message,traceId,retriable}` |
| Postgres adapter | **Stub** | `Open` helper; live repos TBD (pgx) |
| Redis cache | **Done** | go-redis when `REDIS_URL` set; else in-process map |
| Kafka publisher | **Stub** | Logs / buffers; implements `ports.EventPublisher` |
| OpenSearch indexer | **Done** | HTTP `IndexProfile` / `DeleteProfile` when `SEARCH_URL`/`OPENSEARCH_URL` set; wired on provision/update/delete/merge |
| gRPC ProfileService | **Stub** | Hand-written types; proto present, no codegen |
| Observability counters | **Stub** | Process-local atomics |
| OpenAPI 3.1 | **Done** | `api/openapi/profile-v1.yaml` |
| Proto profile.v1 | **Done** | `proto/profile/v1/profile.proto` |
| Docker / compose / Makefile / README | **Done** | |
| SQL migrations | **Done** | `migrations/001`–`016` |
| Zone validation via geofence HTTP | **Partial** | Port wired; memory stub always OK |
| Media-service avatar upload | **Partial** | Port wired; memory store |
| ClickHouse analytics projections | **TODO** | Schema placeholders only |
| Production Kafka client | **TODO** | Stub publisher only |

Legend: **Done** = usable · **Partial** = works for compile/demo · **Stub** = interface only · **TODO** = not started
