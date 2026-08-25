# Prompt 57 — Production Release Candidate

Status of this file at commit time: **gates pending a real GitHub Actions run**. Update the matrix only after Ubuntu execution.

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Commit SHA | *(filled after push)* |
| CI acceptance | *(filled after run)* |
| CI release-candidate | *(filled after run)* |
| Environment | GitHub Actions `ubuntu-latest` (disposable) |
| Traffic profile | k6 `qa/k6/rc_bff.js`: 2→5→12→20 VU, ~70s; assumes 2 vCPU GHA runner + in-memory BFF |

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

Fill **Result** only from Ubuntu logs. Until then: pending.

| Category | Gate | Result | Evidence |
|---|---|---|---|
| Build | 46/46 | pending | CI |
| Unit | 46/46 | pending | CI |
| Race | all | pending | CI |
| Docker | 46/46 | pending | CI |
| Compose | full | pending | CI |
| Migration | full | pending | CI |
| Redis | full | pending | CI |
| Kafka | full | pending | CI |
| Startup | full | pending | CI |
| UI E2E | critical journey | pending | `ci-release-candidate` / `e2e-smoke` |
| Negative E2E | critical failures | pending | `e2e-smoke` + Playwright api |
| Idempotency | critical operations | pending | Go tests + `RC_FULL` HTTP |
| Concurrency | inventory/payment/order | pending | Go tests (100 vs 10 inventory) |
| Load | k6 | pending | `qa/k6/rc_bff.js` in RC job |
| Security | baseline + extended | pending | ZAP + `security-sanity` |
| Accessibility | critical UI | pending | admin login form Playwright |
| Observability | trace journey | pending | `X-Request-Id` echo BFF+catalog |
| Recovery | dependency failures | pending | `recovery-smoke.sh` |
| Mobile store | App Store / Play | BLOCKED | no emulator, no store signing run |
| Flutter customer UI checkout | integration_test | BLOCKED | splash-only `app_boot_test.dart`; no GHA emulator |

## Failures / fixes / reruns

| Run | Result | Notes |
|---|---|---|
| https://github.com/hfzednz/hfzx-getir/actions/runs/32900991522 | FAIL (partial) | `rc-recovery` PASS. `setup-node@v5` failed on journeys/UI (automatic npm cache with no root lockfile). Fixed with `package-manager-cache: false`. |
| https://github.com/hfzednz/hfzx-getir/actions/runs/32903553720 | FAIL (partial) | UI/recovery PASS. Checkout, OMS, payment capture+duplicate refund PASS. Inventory python compared `id` but reservation DTO uses `ID`. |

## Final status

**NOT CERTIFIED** until a real `ci-release-candidate` run on GitHub Actions is green and this matrix is updated from logs.
