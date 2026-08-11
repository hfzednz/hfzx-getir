# Production Validation

Automated entrypoint: `go run ./tools/prod-validate -env=<staging|prod|…>`

## Health matrix

| Layer | Probe | Fail if |
|-------|-------|---------|
| API / BFF | `GET /health` + `/ready` | non-200 or timeout > 2s |
| Identity | ready + JWKS reachable | missing keys |
| Orders | ready + DB ping | not ready |
| Payments | ready; PSP configured (no mock) | mock or missing key |
| Redis | PING | fail |
| Kafka | broker metadata | no brokers |
| OpenSearch | cluster health ≠ red | red |
| AI platform | `/ready` | fail |
| Notifications | provider config present | missing required channel |
| Storage / CDN | HEAD object / edge | 5xx |
| Maps | geocode smoke (location) | upstream 5xx |
| Payment providers | Stripe balance/ping (sandbox vs live) | auth fail |
| Notification providers | Twilio/FCM dry-run | auth fail |

## Business smoke (post canary 10%)

See BUSINESS_VALIDATION.md — customer place, courier accept, warehouse pick, admin flag dual-control, refund path.

## Certification tools

- `tools/integration-cert` — system integration
- `tools/genesis-cert` — autonomy genesis gates
- `tools/prod-validate` — environment health for release
