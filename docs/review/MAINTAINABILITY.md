# Maintainability Report (Prompt-42)

## Improvements

- Single HTTP client implementation for order outbound adapters  
- Explicit production boot failures instead of half-wired stubs  
- Review artifacts under `docs/review/` for continuity  
- Progress track remains `docs/implementation/PRODUCTION_WIRING.md`

## Conventions reinforced

- Money: int64 minor units (unchanged)  
- Errors: camelCase `{error:{code,message,traceId,retriable}}` (unchanged)  
- DevMode: empty `DATABASE_URL`  
- Opaque cross-service IDs (admin courier field is free-text opaque id)

## Debt

- Shared Go packages (`nxpg`, `nxkafka`, `nxratelimit`) still duplicated per service  
- Catalog / payment / inventory incomplete Postgres surfaces  
- Admin feature APIs beyond orders may still carry mock remnants elsewhere  
