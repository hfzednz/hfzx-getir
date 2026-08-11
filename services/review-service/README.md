# Review, Rating, Reputation & Trust Service

NEXORA central trust layer. HTTP `:8103` · gRPC `:9103` · base path `/v1/reviews`.

## Run (dev / memory mode)

```bash
cd services/review-service
go test ./...
go run ./cmd/review-service
```

`DATABASE_URL` empty → in-memory repositories + stub Kafka/OpenSearch/LLM.

## Boundaries

Owns reviews, ratings, moderation, trust, reputation, quality scores.  
Does **not** own catalog, CRM tickets/CSAT, profiles, orders, or media binaries.

## Key headers

- `X-Tenant-Id` (required)
- `X-Nexora-User` (principal)
- `Idempotency-Key` / `X-Request-Id`

## Docs

- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [FEATURES.md](./FEATURES.md)
