# Prompt 58 — Final product acceptance

Status at commit time: **gates added; fill Result only after Ubuntu/macOS GitHub Actions on this commit.** Do not treat local Windows output as PASS.

Customer entrypoint is the **Flutter** app `apps/mobile_customer` (`nexora_customer`). There is no customer Next.js storefront (`admin_web` / `super_admin_web` are ops only).

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Commit SHA | *(filled after push)* |
| CI acceptance | *(filled after run)* |
| CI release-candidate | *(filled after run)* |
| Prompt 57 baseline | `b26e55c` — https://github.com/hfzednz/hfzx-getir/actions/runs/32911312874 |

## Flutter inventory

| App | Path | applicationId / bundle | Version |
|---|---|---|---|
| Customer | `apps/mobile_customer` | Android `com.hfzx.nexora.nexora_customer` / iOS `com.hfzx.nexora.nexoraCustomer` | `1.0.0+1` |
| Courier | `apps/mobile_courier` | `com.nexora.nexora_courier` / `com.nexora.nexoraCourier` | `1.0.0+1` |
| Warehouse | `apps/mobile_warehouse` | `io.nexora.nexora_warehouse` / `io.nexora.nexoraWarehouse` | `1.0.0+1` |

SDK constraint: Dart `^3.5.0`, Flutter `>=3.24.0`. Nested `apps/*/apps/*` trees are accidental `flutter create` scaffolds, not shipping apps.

CD `--dart-define=ENV=prod` is now accepted as `NEXORA_ENV`. Customer `ApiClient` uses `/v1/customer` and sends `X-Tenant-Id`.

## Honesty

- In-memory checkout still uses seeded cart `33333333-3333-3333-3333-333333333333`.
- Full Flutter **emulator** checkout (OPEN APP → every screen) is **not** in GHA (no emulator job). Live gate is `test/live/bff_checkout_journey_test.dart` against the real BFF plus widget failure UI tests.
- k6 numbers remain CI/environment measurements, not production SLA.
- Store listing copy in `store/` is draft. Privacy policy / accounts / screenshots are **EXTERNAL INPUTS** (`store/EXTERNAL_INPUTS.md`).
- Android AAB without Play upload key is debug-signed. iOS job is `--no-codesign`.
- Do not claim App Store / Play approval.

## Matrix

Fill after the Prompt 58 GHA run.

| Gate | Result | Evidence |
|---|---|---|
| Backend acceptance | pending | `ci-acceptance` |
| Race | pending | CI |
| Docker | pending | CI |
| Compose | pending | CI |
| Migrations | pending | CI |
| Redis | pending | CI |
| Kafka | pending | CI |
| Startup | pending | CI |
| Security | pending | CI |
| k6 | pending | `rc-journeys-k6` (not production SLA) |
| Recovery | pending | `rc-recovery` |
| Accessibility | pending | admin `rc-ui-a11y` + customer widget error states |
| Full customer UI checkout | BLOCKED | no GHA emulator; splash `integration_test` only |
| Full customer order journey | pending | HTTP `RC_FULL` + Flutter `test/live` when `FLUTTER_LIVE=1` |
| Flutter customer tests | pending | `rc-flutter-static` |
| Android release build | pending | `rc-android-aab` (signing secrets optional) |
| iOS release build | pending | `rc-ios-build` unsigned; codesign BLOCKED |
| Store readiness | BLOCKED | metadata present; legal URLs / accounts missing |
| CI warnings | pending | setup-go@v7, setup-node@v6, setup-buildx@v4 |
| Mobile performance | BLOCKED | no device profiler in CI |
| Push / associated domains | BLOCKED | no production FCM/APNs secrets in repo |

## Final status

**NOT CERTIFIED** as PRODUCTION PRODUCT CANDIDATE until this matrix is updated from a real `ci-release-candidate` + `ci-acceptance` run on the Prompt 58 commit.
