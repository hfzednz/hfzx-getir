# Prompt 59 — GitHub Actions orchestration diagnostic

Inspected 2026-08-26. Application code and recovery-smoke.sh were **not** changed. No `workflow_dispatch`, no empty commits, no duplicate workflows.

## Identity

| Field | Value |
|---|---|
| Local branch | `main` (`main...origin/main`) |
| `git rev-parse HEAD` | `4420cf7530a9372d08abddb7b2b049f094ee177c` |
| `git ls-remote origin refs/heads/main` | `4420cf7530a9372d08abddb7b2b049f094ee177c` |
| HEAD synchronized | **PASS** |
| Dirty tree (uncommitted) | `PROMPT58_FINAL_PRODUCT_ACCEPTANCE.md` + CRLF `services/*/go.mod` noise. **Not committed** (would retrigger CI and `cancel-in-progress` on RC/Acceptance). |

## Root cause (evidence, not assumption)

**GitHub Actions is in a documented major outage.** The `4420cf7` push landed inside that window.

| Source | Evidence |
|---|---|
| https://www.githubstatus.com/api/v2/status.json | `indicator: major`, `Partial System Outage` (page updated 15:48:07Z) |
| Component `Actions` (`br0l2tvcx85d`) | **`major_outage`** since **2026-08-26T15:11:58Z** |
| Component Pages | `degraded_performance` since 15:12:21Z |
| Incident | https://stspg.io/pg14nv9m3095 — “Incident with Actions”, status **investigating**, impact **critical** |
| Latest incident note (15:48:07Z) | “primary failover briefly improved performance but did not fully mitigate, we've throttled inbound traffic and are investigating upstream Vitess issues” |
| `4420cf7` Quality created | **15:15:30Z** — 3.5 min after Actions went `major_outage` |
| `4420cf7` RC created | **15:16:41Z** |
| `512f553` pipelines | Completed normally **before** the outage (Quality/RC/Acceptance 14:29–14:53Z) |

This is an **external GitHub blocker**. Repository workflow YAML, concurrency, runners, `if:`, environments, and branch filters do not explain 0-job queues: those same files ran jobs on `512f553`.

## `4420cf7` run inventory (API, not fabricated)

| Workflow | Run ID | Event | Status | Jobs | URL |
|---|---|---|---|---|---|
| ci-quality | 32984528739 | push | queued (`updated_at` frozen 15:15:30Z) | **0** | https://github.com/hfzednz/hfzx-getir/actions/runs/32984528739 |
| ci-release-candidate | 32984586382 | push | queued (`updated_at` frozen 15:16:41Z) | **0** | https://github.com/hfzednz/hfzx-getir/actions/runs/32984586382 |
| ci-acceptance | — | — | **not created** (latest still `512f553` run 32980607713) | — | — |

Repo-wide: `queued=2`, `in_progress=0`, `waiting=0`, `pending=0`, `requested=0`. No in-flight run is holding a concurrency slot.

Check suites on `4420cf7` (all `queued`, `latest_check_runs_count=0`): GitHub Actions ×2, Vercel, Cursor — platform-wide stall, not a single workflow `if:`.

## Why this is not a local workflow defect

Checked against live YAML and the GitHub API:

| Hypothesis | Result |
|---|---|
| Concurrency blocking | **FAIL as cause.** Quality has **no** `concurrency`. RC/Acceptance use `cancel-in-progress: true` on `ci-*-${{ github.ref }}`. Previous `512f553` RC/Acceptance **completed** (failure/success) at 14:39 / 14:53. Nothing in-progress to serialize behind. |
| Environment approval | **FAIL as cause.** No job `environment:` on quality / acceptance / RC. |
| Branch / path filters | Quality+RC `on.push.branches: [main]` (RC also `master`). Acceptance push has **no** path filter. `ci-services` / `ci-infra` correctly skipped (`paths:` would not match `scripts/ci/recovery-smoke.sh` only). |
| Job `if:` skipping all jobs | **FAIL as cause.** Quality `nightly-perf-security` is schedule-only; `quality-unit` has no `if`. RC jobs have no `if`. Zero jobs means jobs were **never materialized**, not skipped. |
| Self-hosted / missing labels | **FAIL as cause.** Hosted `ubuntu-latest` (RC iOS is `macos-latest`). Same labels ran at 14:29Z. |
| Workflow disabled | **FAIL as cause.** All 9 workflows `state: active`. |
| Syntax preventing parse | **FAIL as cause.** Runs **were created** for Quality+RC; YAML parsed enough to enqueue. |
| Branch protection / reviewers | Not evidenced. Public repo; same actor `hfzednz` ran `512f553` without a wait-for-approval status. Combined commit status `pending` with `total_count: 0`. |
| Duplicate dispatch | **Not performed.** Equivalent Quality+RC push runs already queued. Dispatch during `major_outage` would add more stuck runs. |

Acceptance missing is consistent with **partial workflow registration during the outage** (Git Operations stayed operational; Actions orchestration did not). It is not explained by acceptance `paths:` (those apply to `pull_request` only).

## Workflow inventory (required CI)

| Workflow | `on.push` | `workflow_dispatch` | Concurrency | Runners | Secrets / env |
|---|---|---|---|---|---|
| ci-quality | `main` (no path filter) | **absent** | none | `ubuntu-latest`; nightly job `if: schedule` | none |
| ci-acceptance | `main, master` | yes | `ci-acceptance-${{ github.ref }}` cancel | `ubuntu-latest` only | `contents: read` |
| ci-release-candidate | `main, master` | yes | `ci-release-candidate-${{ github.ref }}` cancel | ubuntu + **macos-latest** (`rc-ios-build`) | `contents: read`; AAB uses optional keystore secrets inside script |
| ci-services | `main` + `services/**` | no | none | ubuntu | packages write on docker job |
| ci-infra | `main` + `infra/**` | no | none | ubuntu | none |
| cd-mobile / cd-gitops / cd-production-validate | dispatch only | yes | none | ubuntu; cd-mobile iOS `macos-latest` | store / Play secrets on CD only |
| release-changelog | tags `v*` / `mobile/*` | no | none | ubuntu | none |

Concurrency design is intentional serialization per ref with cancel of superseded runs. It is **not** permanently queueing `4420cf7`. Do not remove it.

## Dispatch decision (Phase 9)

| Action | Taken? | Why |
|---|---|---|
| Dispatch Quality | **No** | Push run 32984528739 already queued. Quality has no `workflow_dispatch` anyway. |
| Dispatch Acceptance | **No** | Actions `major_outage`; a new run would likely also stall at 0 jobs. |
| Dispatch RC | **No** | Push run 32984586382 already queued. |

When GitHub marks Actions operational: **do not dispatch** Quality/RC if those two runs are still queued; let them acquire runners. If they are lost/cancelled and Acceptance is still missing, then dispatch Acceptance (and RC if needed) **once**. Do not add `workflow_dispatch` to quality unless a recovered GitHub still cannot start the existing push run.

## Recovery (`4420cf7`)

`rc-recovery` has **not executed**. Do not treat the Kafka-wait change as verified. Do not edit `scripts/ci/recovery-smoke.sh` until a real job log exists.

## Fixes

None in-repo. External: wait for GitHub incident `y1t7p9fzrlj2` / https://www.githubstatus.com (Actions component).

## Matrix at inspection time

Queued ≠ PASS. 0 jobs ≠ PASS.

| Pipeline | Status | Run |
|---|---|---|
| Quality | **BLOCKED** (queued, 0 jobs) | https://github.com/hfzednz/hfzx-getir/actions/runs/32984528739 |
| Acceptance | **BLOCKED** (not created for `4420cf7`) | — |
| Release Candidate | **BLOCKED** (queued, 0 jobs) | https://github.com/hfzednz/hfzx-getir/actions/runs/32984586382 |
| Recovery | **BLOCKED** (job never started) | same RC run |

| Diagnostic | Result |
|---|---|
| HEAD synchronized | **PASS** |
| Workflow trigger | **FAIL** (push registered Quality+RC only; jobs never created; Acceptance never registered) because Actions outage |
| Runner available | **BLOCKED** (GitHub Actions compute/orchestration `major_outage`; hosted labels unchanged) |
| Concurrency | **PASS** (not the blocker) |
| Environment approval | **PASS** (no environment on these workflows) |
| Recovery execution | **BLOCKED** |
