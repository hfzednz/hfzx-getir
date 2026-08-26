# Prompt 58 — Final product acceptance

Filled from GitHub Actions on commit `1eb3b25`. k6 numbers are CI measurements, not a production SLA. Emulator checkout and store signing remain **BLOCKED**.

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Commit SHA | `1eb3b25dadf4ddf966734e5f28fcee111f4ac8fd` |
| CI quality | https://github.com/hfzednz/hfzx-getir/actions/runs/32972893647 (**success**) |
| CI acceptance | https://github.com/hfzednz/hfzx-getir/actions/runs/32972893620 (**success**) |
| CI release-candidate | https://github.com/hfzednz/hfzx-getir/actions/runs/32972893586 (**failure**) |

Times (UTC): quality 13:13:09–13:13:38; RC 13:13:09–13:25:07; acceptance 13:13:09–13:38:36.

## ci-quality — PASS

| Job | Result | Steps |
|---|---|---|
| quality-unit | PASS | checkout, setup-go, Quality service tests, Integration cert tool |
| quality-gates | PASS | k6 scripts present, chaos manifests, Playwright config |
| nightly-perf-security | SKIPPED | not nightly |

Artifacts: none.

## ci-acceptance — PASS

| Job | Result |
|---|---|
| go-build-test-verify | PASS |
| go-race-all | PASS |
| docker-build-all | PASS |
| compose-migration-smoke | PASS |
| e2e-smoke | PASS |
| security-sanity | PASS |
| service-startup-smoke | PASS |

All listed steps on those jobs succeeded. Artifacts: none.

## ci-release-candidate — FAIL (`rc-flutter-static` only)

| Job | Result | Notes |
|---|---|---|
| rc-ui-a11y | PASS | admin HTML + a11y |
| rc-recovery | PASS | Redis/Kafka restart |
| rc-journeys-k6 | PASS | k6 + ZAP + Flutter live BFF on this commit (place/idempotency; GET 400 accepted) |
| rc-android-aab | PASS | debug-signed AAB uploaded |
| rc-ios-build | PASS | unsigned `--no-codesign` |
| rc-flutter-static | FAIL | step `Flutter pub get / analyze / test (3 apps)` ~10m50s, exit 1 |

Artifact: `customer-release-aab` (79.6 MB, sha256 `7ea3d193465df613de20e640d36594b6866b759847573c3f893416df682d265c`). Debug-signed. **Not** store signing.

### rc-flutter-static failure

| Field | Value |
|---|---|
| WORKFLOW | ci-release-candidate |
| JOB | rc-flutter-static |
| STEP | Flutter pub get / analyze / test (3 apps) |
| ERROR | Process completed with exit code 1; job wall ~11m54s (test step ~10m50s) |
| ROOT CAUSE | `home_smoke_test` mounted `NexoraApp` + `bootstrap()` / splash `postBootstrap` (Hive/Firebase/FCM/sync). Same 10-minute `TimeoutException` was logged on `2fd4cb5`. Credential fill timed out this run so the 1eb3b25 log body was not re-downloaded; duration matches that timeout. |
| AFFECTED FILE | `apps/mobile_customer/test/widget/home_smoke_test.dart` |

Follow-up (not yet on this SHA): stop mounting `NexoraApp` in the VM smoke test; `flutter test --timeout 30s`. Local `home_smoke_test` **PASS**.

## Matrix (`1eb3b25`)

| Gate | Result |
|---|---|
| Backend | PASS |
| Race | PASS |
| Docker | PASS |
| Compose | PASS |
| Migrations | PASS |
| Redis | PASS |
| Kafka | PASS |
| Startup | PASS |
| Security | PASS |
| k6 | PASS (CI measurement, not production SLA) |
| Recovery | PASS |
| Accessibility | PASS (admin) |
| Flutter tests | FAIL |
| Android AAB | PASS (unsigned/debug-signed artifact) |
| iOS no-codesign | PASS |
| Full emulator checkout | BLOCKED |
| Store signing | BLOCKED |
| Live BFF | PASS (inside `rc-journeys-k6`) |

## Final status

**NOT CERTIFIED** as PRODUCTION PRODUCT CANDIDATE. Quality and acceptance are green. RC is red solely on Flutter static tests. Emulator checkout and store signing stay BLOCKED.
