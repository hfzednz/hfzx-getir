# Prompt 58 — Final product acceptance

Filled from GitHub Actions on `8f2a224`. Emulator checkout and store signing remain **BLOCKED**. k6 numbers are CI measurements, not a production SLA.

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Commit SHA | `8f2a224` |
| CI acceptance | https://github.com/hfzednz/hfzx-getir/actions/runs/32959142652 (**success**) |
| CI release-candidate | https://github.com/hfzednz/hfzx-getir/actions/runs/32959142609 (**failure**) |
| CI quality | https://github.com/hfzednz/hfzx-getir/actions/runs/32959142599 (**failure**) |
| Prompt 57 baseline | `b26e55c` — https://github.com/hfzednz/hfzx-getir/actions/runs/32911312874 |

## Job results (`8f2a224`)

### ci-acceptance — PASS

| Job | Result |
|---|---|
| go-build-test-verify | PASS |
| go-race-all | PASS |
| docker-build-all | PASS |
| compose-migration-smoke | PASS |
| service-startup-smoke | PASS |
| security-sanity | PASS |
| e2e-smoke | PASS |

### ci-release-candidate — FAIL

| Job | Result | Notes |
|---|---|---|
| rc-ui-a11y | PASS | admin HTML + a11y |
| rc-recovery | PASS | Redis/Kafka restart |
| rc-journeys-k6 | FAIL | step “Full journeys + idempotency + k6 + ZAP + Flutter live BFF”; log body not downloadable without auth in this session. k6 not claimed PASS. |
| rc-flutter-static | FAIL | `flutter-static.sh` (~23s after Flutter setup) |
| rc-android-aab | FAIL | AAB step ~99s; also annotation: migrate `setup-java` v4 → v5 (Node 20 deprecation) |
| rc-ios-build | FAIL | `flutter build ios --no-codesign` ~2m; **not** a signed IPA |

### ci-quality — FAIL

| Job | Result | Evidence |
|---|---|---|
| quality-unit | FAIL | Failed in **0s**. `services/quality-service/go.mod` requires **go 1.25.0**; workflow pinned **1.22.x**. Same job was green on `65a4020` with `setup-go@v5`. `setup-go@v7` no longer silently upgrades the toolchain. Cache warning: root `go.mod` missing. |
| quality-gates | skipped | needs quality-unit |

## Matrix

| Gate | Result | Evidence |
|---|---|---|
| Backend acceptance | PASS | https://github.com/hfzednz/hfzx-getir/actions/runs/32959142652 |
| Race | PASS | `go-race-all` |
| Docker | PASS | `docker-build-all` |
| Compose | PASS | `compose-migration-smoke` |
| Migrations | PASS | `compose-migration-smoke` |
| Redis | PASS | `compose-migration-smoke` |
| Kafka | PASS | `compose-migration-smoke` |
| Startup | PASS | `service-startup-smoke` |
| Security | PASS | `security-sanity` |
| k6 | FAIL | `rc-journeys-k6` failed; no authenticated log excerpt in this pass |
| Recovery | PASS | `rc-recovery` |
| Accessibility | PASS | `rc-ui-a11y` (admin). Customer widget tests did not run (`rc-flutter-static` FAIL) |
| Full customer UI checkout | BLOCKED | no emulator/device job executed |
| Full customer order journey | FAIL | live Flutter BFF is inside failed `rc-journeys-k6` |
| Flutter customer tests | FAIL | `rc-flutter-static` |
| Android release build | FAIL | `rc-android-aab`; unsigned/debug-signed store upload still BLOCKED |
| iOS release build | FAIL | unsigned compile failed; codesign BLOCKED |
| Store readiness | BLOCKED | `store/EXTERNAL_INPUTS.md`; no signing credentials |
| CI warnings | WARN | `setup-java@v4` Node 20 (to be upgraded); quality checkout@v4 Node 20 |
| Mobile performance | BLOCKED | no device profiler |
| Push / associated domains | BLOCKED | no production FCM/APNs secrets |

## Follow-up fix (verified only)

- `ci-quality.yml`: `go-version: 1.26.x` + `cache-dependency-path` for the quality module (matches go.mod 1.25+ and acceptance).
- `rc-android-aab`: `actions/setup-java@v5` per the failing run’s annotation.

Flutter/Android/iOS/journeys script changes wait on **job log text**, not guesses.

## Final status

**NOT CERTIFIED** as PRODUCTION PRODUCT CANDIDATE. 46-service acceptance is green on `8f2a224`. Product RC jobs for Flutter/AAB/iOS/live BFF are red. Emulator checkout and store signing stay BLOCKED.
