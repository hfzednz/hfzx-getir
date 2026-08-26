# Prompt 61 — Recover stale queues; fresh CI on current code

Stale outage runs are **invalid evidence**. Fresh runs triggered by legitimate documentation push after Actions recovery.

## Identity

| Field | Value |
|---|---|
| Repository | https://github.com/hfzednz/hfzx-getir |
| Branch | main |
| Code under test | recovery fix on `4420cf7530a9372d08abddb7b2b049f094ee177c` (`scripts/ci/recovery-smoke.sh` Kafka broker API wait) |
| GitHub Status | https://www.githubstatus.com — operational at trigger time |

## Stale runs (invalid — do not cite as PASS)

Created during Actions `major_outage` (2026-08-26T15:11:58Z). Remained `queued` with **0 jobs** after recovery; `updated_at` frozen.

| Workflow | Run ID | URL | Why invalid |
|---|---|---|---|
| ci-quality | 32984528739 | https://github.com/hfzednz/hfzx-getir/actions/runs/32984528739 | 0 jobs, no runner, no logs |
| ci-release-candidate | 32984586382 | https://github.com/hfzednz/hfzx-getir/actions/runs/32984586382 | 0 jobs, no runner, no logs |
| ci-acceptance | — | — | never created for `4420cf7` |

## Fresh trigger

Legitimate docs commit (PROMPT59/60/61) pushed to `main` after Actions recovery. No application or recovery logic changes. No empty commit. `ci-quality` has no `workflow_dispatch`; push is the required trigger for Quality + Acceptance + RC together.

*(Fresh run URLs, SHA, job table, recovery logs filled below after execution.)*

## Fresh runs

| Workflow | Run ID | SHA | Status | URL |
|---|---|---|---|---|
| ci-quality | TBD | TBD | TBD | TBD |
| ci-acceptance | TBD | TBD | TBD | TBD |
| ci-release-candidate | TBD | TBD | TBD | TBD |

## rc-recovery

**BLOCKED** until fresh `rc-recovery` job starts and logs are captured.

*(Recovery log evidence appended after fresh RC completes.)*

## Matrix

| Pipeline | Status | Evidence |
|---|---|---|
| Fresh Quality | TBD | TBD |
| Fresh Acceptance | TBD | TBD |
| Fresh RC | TBD | TBD |
| rc-recovery | TBD | TBD |

| Condition | Status |
|---|---|
| HEAD synchronized | PASS (pre-push `4420cf7`) |
| Actions operational | PASS |
| Fresh runner acquired | TBD |
| Acceptance executed | TBD |
| RC executed | TBD |
| Recovery executed | TBD |
| Recovery logs available | TBD |
| Recovery assertions | TBD |
