# Prompt-46 missing-file completion

Generated only previously missing production implementations (architecture frozen).

| Area | Files |
|------|--------|
| payment-service | postgres Intent/Outbox; HTTP fraud/wallet/ledger clients; config URLs; `.env.example` |
| autonomy-service | postgres repos swap + `.env.example` |
| inventory-service | advisory lock, idempotency store, migration `015`, `.env.example` |
| cart / checkout | real HTTP clients + checkout complete lock + `.env.example` |
| wallet-service | postgres WalletRepo + OutboxRepo + main wiring + `.env.example` |
| chat-assistant-service | CRM proxy facade `:8126`, Dockerfile, `.env.example` |
| geofence-service | real postgres `Open` |
| warehouse-service | real inventory HTTP client |
| catalog-service | suppliers port/memory/postgres/use-cases/HTTP |
| pricing-service | real promo HTTP Evaluate client |
| settlement-service | postgres Batch/Event/Outbox + main |
| finance-ledger-service | postgres Account/Journal/Invoice/Tax/Event/Outbox + main |
| promotion-service | postgres all repos + main swap |
| pricing-service | real promo HTTP Evaluate client |
| geofence-service | postgres ZoneRepo + OutboxRepo + main swap + `.env.example` |

Verification: `go test ./...` on touched services.
