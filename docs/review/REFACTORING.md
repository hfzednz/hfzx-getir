# Refactoring Report (Prompt-42)

## Applied

| Change | Why |
|--------|-----|
| order-service `adapters/{inventory,payment,warehouse,dispatch}` now wrap `httpclients` | Removed stub-vs-HTTP duplication; single HTTP implementation |
| order `main` uses named adapter packages + `AdvisoryPlaceLocker` | Clearer DI; distributed place lock when Postgres is on |
| bff-customer OTP split into `StartOTP` + `VerifyOTP` | Matches identity-service challenge flow |
| Checkout `Place` accepts `sessionId` from preview | Avoids discarding preview session |
| admin_web orders `api.ts` | Removed unused mock builders (~300 LOC dead code) |

## Deferred (no behavior change required yet)

- Per-service duplicated `ratelimit` packages — leave until shared Go module lands  
- Catalog half-swap (products PG, rest memory) — logged; full swap needs adapters  
