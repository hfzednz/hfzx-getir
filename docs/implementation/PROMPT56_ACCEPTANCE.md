# Prompt 56A — GitHub Remote, Push & CI Evidence

**Date:** 2026-08-18  
**Completion state:** **STATE A — VERIFIED COMPLETE** for `ci-acceptance` required Linux gates.

## Repository

| Field | Value |
|-------|--------|
| Local root | `hfzx_Getir_` (`git rev-parse --show-toplevel` is the product directory) |
| Remote | `https://github.com/hfzednz/hfzx-getir.git` |
| Branch | `main` |
| Workflow | `ci-acceptance` (`.github/workflows/ci-acceptance.yml`) |
| Trigger | `push` to `main` (`workflow_dispatch` is also defined) |

## Verified green run

| Field | Value |
|-------|--------|
| Run ID | `32133592912` |
| Run URL | https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912 |
| Commit SHA | `b114014705fa6943d6266029b6e3c2917c8f907d` |
| Commit | `fix(ci): wait for Postgres before startup smoke probes` |
| Event | `push` |
| Started | 2026-08-18T11:48:49Z |
| Completed | 2026-08-18T12:13:33Z |
| Conclusion | **success** |

## Job results (this run)

| Job | Status | Window (UTC) | URL |
|-----|--------|--------------|-----|
| go-build-test-verify | **success** | 11:48:52–11:51:04 | [job](https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912/job/95699833666) |
| go-race-all | **success** | 11:51:06–11:53:39 | [job](https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912/job/95700405020) |
| compose-migration-smoke | **success** | 11:51:06–11:52:24 | [job](https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912/job/95700404997) |
| docker-build-all | **success** | 11:51:06–12:07:41 | [job](https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912/job/95700405019) |
| service-startup-smoke | **success** | 12:07:43–12:13:33 | [job](https://github.com/hfzednz/hfzx-getir/actions/runs/32133592912/job/95704779320) |

Same SHA also ran **ci-quality** successfully: https://github.com/hfzednz/hfzx-getir/actions/runs/32133592895 (unit + file-presence gates; not ZAP/k6).

## Failures repaired before green

### Run 32129202206 (`9ca4bbc`) — compose-migration-smoke FAIL

| Item | Detail |
|------|--------|
| JOB | compose-migration-smoke |
| STEP | Migration + Redis + Kafka smoke |
| COMMAND | `bash scripts/ci/migration-smoke.sh` → `docker compose up -d postgres redis kafka` |
| ERROR | `manifest for bitnami/kafka:3.7 not found: manifest unknown` |
| ROOT CAUSE | Public `bitnami/kafka` tags were moved off Docker Hub |
| FILE | `infra/docker/docker-compose.yml` (and matching service compose files) |
| FIX | `bitnamilegacy/kafka:3.7` (same Bitnami env/paths) |
| RERUN | push `bfc91e2` |

### Run 32129963216 (`bfc91e2`) — service-startup-smoke FAIL

| Item | Detail |
|------|--------|
| JOB | service-startup-smoke |
| STEP | Core service startup + health + SIGTERM |
| COMMAND | `bash scripts/ci/service-startup-smoke.sh` |
| ERROR | Job failed in **27s** (too fast for image builds); script called `pg_isready` immediately after `up -d` |
| ROOT CAUSE | No Postgres readiness wait (unlike `migration-smoke.sh`) |
| FILE | `scripts/ci/service-startup-smoke.sh` |
| FIX | Wait loop (up to ~120s) then `pg_isready` |
| RERUN | push `b114014` → **success** |

## Acceptance matrix

| Gate | Status | Evidence |
| ---- | ------ | -------- |
| 46/46 tests | **PASS** | `go-build-test-verify` success on run `32133592912` |
| 46/46 builds | **PASS** | same job (`go build ./...` per service) |
| go mod verify | **PASS** | same job (`go mod verify` per service) |
| race | **PASS** | `go-race-all` success |
| Docker 46/46 | **PASS** | `docker-build-all` success; 46 service Dockerfiles in tree |
| Compose | **PASS** | compose config validate + `compose-migration-smoke` success |
| migrations | **PASS** | `compose-migration-smoke` success |
| Redis | **PASS** | same job (Redis PING + SET/GET smoke) |
| Kafka | **PASS** | same job (broker ready + topic smoke) |
| startup | **PASS** | `service-startup-smoke` success |
| E2E | **BLOCKED** | No Playwright/E2E job in `ci-acceptance.yml`; tests were not executed |
| security | **BLOCKED** | No ZAP/security job in `ci-acceptance.yml`; `nightly-perf-security` is schedule-only and did not run |

## Honesty

Gates marked **PASS** were observed as GitHub Actions `conclusion=success` on Ubuntu for run `32133592912`. E2E and security remain **BLOCKED** because those jobs did not execute.
