# Production implementation progress (post-#40 / Prompt-46+)

Additive production wiring only — architecture frozen.

## Completed

| Component | Production path |
|-----------|-----------------|
| Core commerce (identity/order/payment/wallet/inventory/cart/checkout/catalog) | Postgres + HTTP clients |
| Money/fulfillment (settlement/ledger/promo/pricing/loyalty/notification/dispatch/warehouse) | Postgres repos + clients |
| Profile/CRM/geofence/location/tracking/routing/review/recommendation | Postgres (+ location Redis GEO) |
| Platform satellites (ai/search/security/liveops/platform-ops/quality/superapp/open-platform/hyperscale) | Postgres repos + main swap |
| `erp-service` / `data-platform-service` | Postgres all repos + main swap |
| `supplier-service` | Postgres all sourcing/marketplace repos + main swap |
| `global-service` / `innovation-service` / `enterprise-ops-service` | Postgres repos + main swap |
| `pricing → inventory` | DynamicHintClient HTTP ATP via `INVENTORY_BASE_URL` |
| `catalog → OpenSearch` | HTTP `_doc` index/search/suggest when `OPENSEARCH_URL` set |
| `order → OpenSearch` | HTTP `_doc` index/delete/search on `nexora-orders` when `OPENSEARCH_URL` set (in-process mem fallback) |
| `catalog → media` | HTTP GET `/v1/media/assets/{id}` when `MEDIA_SERVICE_URL` set (synthesized CDN fallback on failure) |
| `loyalty → wallet` | HTTP ensure+credit when `WALLET_BASE_URL` set (synthetic success when empty) |
| `settlement → payout` | HTTP POST payout provider when `PAYOUT_PROVIDER_URL` set (synthetic success when empty) |
| `catalog` / `promotion` Redis | Real go-redis Ping/Get/Set when `REDIS_URL` set |
| `review → OpenSearch` | HTTP review index/search when `OPENSEARCH_URL` set |
| `location → OpenSearch` | geo_point dual-write POI/address indexer when `OPENSEARCH_URL` set |
| `settlement` / `finance-ledger` Redis | Real go-redis Ping when `REDIS_URL` set |
| `geofence` / `dispatch` / `crm` Redis | Real go-redis Open+Ping when `REDIS_URL` set |
| `customer-profile` Redis cache | Real go-redis Get/Set with in-process fallback |
| `order` / `cart` / `checkout` / `loyalty` / `notification` / `pricing` Redis | Real go-redis Ping when `REDIS_URL` set |
| `warehouse` Redis queue | Real LPUSH/RPOP task queue when `REDIS_URL` set |
| `checkout → promo/fraud/geofence` | HTTP when `PROMO_URL`, `FRAUD_URL`/`AI_PLATFORM_URL`, `GEOFENCE_URL` set |
| `settlement → finance-ledger` | HTTP journals when `LEDGER_BASE_URL` set |
| `routing → weather` | HTTP weather factor when `WEATHER_URL`/`OPENWEATHER_URL` set |
| `customer-profile → OpenSearch` | HTTP indexer when `SEARCH_URL`/`OPENSEARCH_URL` set |
| `.env.example` catalog/promotion | Synced with config keys |
| Kafka publishers (segmentio) | Real WriteMessages when `KAFKA_BROKERS` set; noop when empty |

## Still converting

- Intentional gRPC listeners where REST is primary (codegen deferred).
- Kafka empty-broker noop is intentional for local/dev; non-empty `KAFKA_BROKERS` uses real segmentio/kafka-go writers across upgraded services.

DevMode (`DATABASE_URL` empty) remains intentional for local unit tests.

See: `docs/implementation/PROMPT46_COMPLETION.md` · Omega: `docs/implementation/OMEGA_CERTIFICATION.md`
