# Prompt 61 — Recover stale queues; fresh CI on current code

Stale outage runs are **invalid evidence**. Fresh runs triggered by legitimate documentation push (`3181791`) after GitHub Actions recovery. Application and recovery logic unchanged on this push (docs-only delta from `4420cf7`).

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Code under test | `4420cf7530a9372d08abddb7b2b049f094ee177c` recovery fix (`scripts/ci/recovery-smoke.sh` Kafka broker API wait) |
| Fresh trigger commit | `318179146c7f5179fb2040e920bf4e54a52e15d1` (docs only) |
| GitHub Status at trigger | https://www.githubstatus.com — All Systems Operational |

## Stale runs (invalid — do not cite as PASS)

Created during Actions `major_outage` (2026-08-26T15:11:58Z). Remained `queued` with **0 jobs** after recovery; `updated_at` frozen.

| Workflow | Run ID | URL | Why invalid |
|---|---|---|---|
| ci-quality | 32984528739 | https://github.com/hfzednz/hfzx-getir/actions/runs/32984528739 | 0 jobs, no runner, no logs |
| ci-release-candidate | 32984586382 | https://github.com/hfzednz/hfzx-getir/actions/runs/32984586382 | 0 jobs, no runner, no logs |
| ci-acceptance | — | — | never created for `4420cf7` |

## Fresh trigger method

Legitimate docs commit (PROMPT59/60/61) pushed to `main` after Actions recovery. No empty commit. No `workflow_dispatch` (push triggers Quality + Acceptance + RC; `ci-quality` has no dispatch). Stale runs left untouched.

## Fresh runs — PASS (2026-08-26)

| Workflow | Run ID | SHA | Result | Duration (UTC) | URL |
|---|---|---|---|---|---|
| ci-quality | 33002304454 | `3181791` | **success** | 18:54:50–18:55:39 | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304454 |
| ci-acceptance | 33002304601 | `3181791` | **success** | 18:54:50–19:19:58 | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304601 |
| ci-release-candidate | 33002304473 | `3181791` | **success** | 18:54:50–19:04:07 | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304473 |

### ci-quality jobs

| Job | Result |
|---|---|
| quality-unit | PASS |
| quality-gates | PASS |
| nightly-perf-security | SKIPPED |

Artifacts: none.

### ci-acceptance jobs

| Job | Result | Window (UTC) |
|---|---|---|
| go-build-test-verify | PASS | 18:54:53–18:56:08 |
| go-race-all | PASS | 18:56:10–18:58:48 |
| compose-migration-smoke | PASS | 18:56:10–18:57:25 |
| e2e-smoke | PASS | 18:56:10–18:59:37 |
| security-sanity | PASS | 18:56:11–18:57:57 |
| docker-build-all | PASS | 18:56:10–19:14:07 |
| service-startup-smoke | PASS | 19:14:09–19:19:57 |

Artifacts: none.

### ci-release-candidate jobs

| Job | Result | Runner |
|---|---|---|
| rc-flutter-static | PASS | ubuntu-latest |
| rc-ui-a11y | PASS | ubuntu-latest |
| rc-recovery | PASS | ubuntu-latest |
| rc-journeys-k6 | PASS | ubuntu-latest |
| rc-android-aab | PASS | ubuntu-latest |
| rc-ios-build | PASS | macos-latest |

Artifact: `customer-release-aab` (79.6 MB, sha256 `7bbd26327d12f9a3cd42cc53c1d4cd703393d2c325497e1ab935da141670b94c`). Debug-signed. **Not** store signing.

## rc-recovery — PASS (verified logs)

| Field | Value |
|---|---|
| Run | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304473 |
| Job | rc-recovery (job id 98287009030) |
| Step | Redis/Kafka restart + Postgres integrity |
| Duration | 18:54:53–18:55:19Z (~26s) |
| Exit | success |

Log evidence (downloaded from job 98287009030):

1. Infra up: postgres, redis, kafka
2. `==> wait postgres` → accepting connections
3. `==> wait redis` → `OK redis`
4. `==> restart redis` → wait → `OK redis`
5. `==> postgres still ready after redis restart` → accepting connections
6. `==> restart kafka` → `==> wait kafka broker API` → `OK kafka broker API`
7. `OK postgres data plane intact after kafka restart`
8. **`RECOVERY_SMOKE_PASS`**

The `4420cf7` fix (broker API wait instead of `sleep 8` + compose-ps grep) validated on real Ubuntu runner.

## Matrix

| Pipeline | Status | Evidence |
|---|---|---|
| Fresh Quality | **PASS** | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304454 |
| Fresh Acceptance | **PASS** | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304601 |
| Fresh RC | **PASS** | https://github.com/hfzednz/hfzx-getir/actions/runs/33002304473 |
| rc-recovery | **PASS** | job 98287009030; `RECOVERY_SMOKE_PASS` in logs |

| Condition | Status |
|---|---|
| HEAD synchronized | **PASS** (`3181791` on origin/main) |
| Actions operational | **PASS** |
| Fresh runner acquired | **PASS** |
| Acceptance executed | **PASS** |
| RC executed | **PASS** |
| Recovery executed | **PASS** |
| Recovery logs available | **PASS** |
| Recovery assertions | **PASS** (`RECOVERY_SMOKE_PASS`) |

## External gates (unchanged)

| Gate | Status |
|---|---|
| Full emulator checkout | **BLOCKED** |
| Production store signing | **BLOCKED** |
| App Store submission | **BLOCKED** |
| Google Play submission | **BLOCKED** |

k6 numbers in RC are CI measurements, not production SLA.
