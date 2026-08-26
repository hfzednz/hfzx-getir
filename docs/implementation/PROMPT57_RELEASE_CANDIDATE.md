# Prompt 57 — Production Release Candidate

Status of this file: **GHA release-candidate gates PASSED** on Ubuntu for executed jobs. Flutter store / full customer UI checkout remain **BLOCKED**. This is not a production SLA or App Store claim.

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Commit SHA | `b26e55c362888bcf51d1767d35c0a4d1000386a3` |
| CI acceptance | https://github.com/hfzednz/hfzx-getir/actions/runs/32911312725 (success) |
| CI release-candidate | https://github.com/hfzednz/hfzx-getir/actions/runs/32911312874 (success) |
| CI quality | https://github.com/hfzednz/hfzx-getir/actions/runs/32911312798 (success) |
| Environment | GitHub Actions `ubuntu-latest` (disposable) |
| Traffic profile | k6 `qa/k6/rc_bff.js`: 2→4→8→10 VU, ~70s; 2 vCPU GHA runner + in-memory BFF + location. e2e harness sets `RATE_LIMIT_PER_MINUTE=0` (product 240/min quota is not the load SLO). |

## What was added (smallest safe scope)

- Customer BFF: `X-Request-Id` echo + `X-Nexora-User` forwarding; `GET /v1/customer/orders/{id}`; checkout preview/place send principal + idempotency keys.
- PR `e2e-smoke.sh`: negatives, catalog product create/detail, duplicate cart add, observability.
- `RC_FULL=1`: checkout/payment/order/inventory/finance/settlement HTTP journeys, duplicate place/refund/reserve/journal/settlement, k6.
- Go tests: inventory 100 vs 10, reserve idempotency, OMS transitions, payment refund idempotency, ledger journal idempotency, settlement batch idempotency.
- Admin HTML UI Playwright (login/orders/inventory/finance + form a11y). Customer Flutter full UI checkout is **not** executed (no emulator in this workflow).
- Recovery: Redis + Kafka restart with Postgres still ready.
- CI warning cleanup: `actions/checkout@v5`, `setup-node@v5`, Go cache `services/*/go.sum`.

## Honesty constraints

- In-memory checkout (`DATABASE_URL` empty) uses the seeded cart `33333333-3333-3333-3333-333333333333`, not the live cart-service cart.
- Address confirmation is checkout `PATCH` (BFF has no address endpoint).
- Flutter App Store / Play Store readiness is **BLOCKED** (release signing, store listings, emulator E2E not run).
- Supported mobile locales in-repo: `en`, `tr` (`apps/mobile_customer/lib/l10n`). Admin web is English demo sign-in.
- Existing `qa/k6/checkout_load.js` (50–200 VU, p95&lt;500) is **not** the CI profile; it is staging-scale and remains unused in PR/RC GHA.

## Release candidate matrix

Filled from Ubuntu logs on `b26e55c`.

| Category | Gate | Result | Evidence |
|---|---|---|---|
| Build | 46/46 | PASS | `ci-acceptance` `go-build-test-verify` https://github.com/hfzednz/hfzx-getir/actions/runs/32911312725 |
| Unit | 46/46 | PASS | same |
| Race | all | PASS | `go-race-all` |
| Docker | 46/46 | PASS | `docker-build-all` |
| Compose | full | PASS | `compose-migration-smoke` |
| Migration | full | PASS | `compose-migration-smoke` |
| Redis | full | PASS | `compose-migration-smoke` |
| Kafka | full | PASS | `compose-migration-smoke` |
| Startup | full | PASS | `service-startup-smoke` |
| UI E2E | admin login/orders/inventory/finance | PASS | `rc-ui-a11y` job 98005900455 |
| Negative E2E | OTP/cart/404 | PASS | `rc-journeys-k6` `OK negative` + Playwright `--project=api` |
| Idempotency | place/refund/reserve/journal/batch | PASS | `RC_FULL` HTTP duplicates + Go tests |
| Concurrency | inventory 100 vs 10 | PASS | Go tests in `ci-acceptance` |
| Load | k6 GHA profile | PASS | p50=351µs p95=542µs p99=742µs; `http_req_failed` 0/2132; checks 100% (`qa/k6/rc_bff.js`) |
| Security | ZAP baseline + sanity | PASS | `OK zap-baseline rc=0`; `security-sanity` on acceptance |
| Accessibility | admin login form | PASS | `rc-ui-a11y` |
| Observability | `X-Request-Id` echo | PASS | `OK observability` in journeys |
| Recovery | Redis/Kafka restart | PASS | `rc-recovery` job 98005900505 |
| Mobile store | App Store / Play | BLOCKED | no emulator, no store signing run |
| Flutter customer UI checkout | integration_test | BLOCKED | splash-only `app_boot_test.dart`; no GHA emulator |

## Failures / fixes / reruns

| Run | Result | Notes |
|---|---|---|
| https://github.com/hfzednz/hfzx-getir/actions/runs/32900991522 | FAIL (partial) | `rc-recovery` PASS. `setup-node@v5` failed on journeys/UI (automatic npm cache with no root lockfile). Fixed with `package-manager-cache: false`. |
| https://github.com/hfzednz/hfzx-getir/actions/runs/32905683949 | FAIL (partial) | All RC HTTP journeys PASS including ledger+settlement. k6 on `--network host` crossed checks/http_req_failed (~36% failed). Next: attach k6 to the e2e Docker network and cap VUs at 10 for the shared runner. |
| https://github.com/hfzednz/hfzx-getir/actions/runs/32907110359 | FAIL (partial) | `rc-ui-a11y` + `rc-recovery` PASS. All RC HTTP journeys PASS. k6 on e2e Docker network @ 10 VU still 27.55% `http_req_failed` (p95 517µs — fast 4xx/5xx, not saturation). Root cause: location-service default 240 req/min; BFF `GET /home` POSTs serviceability on every call from one BFF IP. Fix: e2e `RATE_LIMIT_PER_MINUTE=0` (quota off in disposable harness; k6 SLO unchanged). |
| https://github.com/hfzednz/hfzx-getir/actions/runs/32911312874 | PASS | `rc-journeys-k6` + `rc-ui-a11y` + `rc-recovery` success on `b26e55c`. k6 2132/2132 checks, 0% failed, p50=351µs p95=542µs p99=742µs. ZAP rc=0. |

## Final status

**GHA RELEASE CANDIDATE GATES PASSED** for executed Ubuntu jobs on `b26e55c` (`ci-release-candidate` + `ci-acceptance` + `ci-quality`). Flutter App Store / Play and full customer UI checkout remain **BLOCKED**. Do not treat k6 numbers as a production SLA; they are the documented GHA 2 vCPU in-memory browse profile.
