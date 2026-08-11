# Bug-Fix Report (Prompt-42)

## P0 — fixed

1. **Payment MockPSP failover** — Stripe outage previously authorized via mock. Non-dev / Stripe path now uses Failover(stripe) only; MockPSP only in DevMode.  
2. **Identity JWT ephemeral key with DB** — `JWT_KEY_PEM` required when `DATABASE_URL` set.  
3. **Identity OTP pepper / OTP_DEV_MODE** — pepper required in prod; known default rejected; `OTP_DEV_MODE` defaults false when DB set and must stay false.  
4. **CORS `*` with DB** — rejected at config load.

## P1 — fixed

5. **Order optimistic lock** — `version <= $new` → `version = $old` (`expectedVersion = o.Version-1`).  
6. **PlaceOrder saga conflict** — `GetByID` error no longer swallowed.  
7. **PlaceLock** — Postgres advisory lock when DB wired.  
8. **BFF OTP login** — no longer starts a new challenge then verifies client code against it.  
9. **BFF KillSwitch fail-open** — returns error when LiveOps missing/fails.  
10. **Authorize soft idempotency** — mismatched key → `ErrConflict`.  
11. **Payment false-ready DB** — stub `Open(nil,nil)` removed; service refuses `DATABASE_URL` until repos exist.  
12. **admin_web `cr_214` / 50% refund** — courier ID and refund minor units required from UI.  
13. **Catalog / identity shutdown** — close publisher + DB handles.

## Still open

- Incomplete domain Postgres coverage (catalog beyond products, payment, inventory, …)  
- Checkout Place eligibility when reusing session without amount (eligibility may omit amountMinor)  
