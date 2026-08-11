# Prompt 53 — CI Execution & Acceptance Closure

**Date:** 2026-08-11  
**Environment (agent):** Windows 10 · Go 1.26.5 · CGO=0 · **no gcc** · **no Docker** · **no `gh`** · **no `act`**

## Critical honesty rule

**No Linux CI run was executed from this environment.**  
Gates that require a GitHub-hosted `ubuntu-latest` runner remain **BLOCKED** until `ci-acceptance.yml` is triggered on GitHub.

### How to obtain PASS evidence

1. Push these files to a GitHub remote (or open a PR touching `services/**` / `scripts/ci/**` / the workflow).
2. Or: Actions → **ci-acceptance** → **Run workflow** (`workflow_dispatch`).
3. Attach the run URL + job logs as evidence, then update this matrix from BLOCKED → PASS/FAIL.

## What was fixed for Linux executability (static / preemptive)

| Change | Why |
|--------|-----|
| `ci-acceptance.yml` → Go **1.26.x** | Modules require go 1.23–1.26; 1.23.x would fail BFFs |
| `permissions: contents: read` | Least privilege |
| `chmod +x scripts/ci/*.sh` | Windows checkouts lose executable bit |
| `GOTOOLCHAIN=auto` | Toolchain download if needed |
| `docker-build-all.sh` temp Dockerfile bump `golang:1.22→1.26` | Avoid 46 permanent Dockerfile rewrites; still builds |
| `DOCKER_BUILDKIT=1` | Layer caching |
| `migration-smoke.sh` discovers **all** `services/*/migrations` | Not a hard-coded 8 |
| Kafka/Redis smoke **hard-fail** | No `|| echo WARN` success hide |
| Compose **healthchecks** for redis + kafka | Readiness without fake wait |
| `service-startup-smoke.sh` | Core subset: build/run/health/SIGTERM |
| `go mod verify` in build job | Dependency integrity |

## Classification

| Kind | Count | Notes |
|------|-------|-------|
| Containerized Go services | 46 | `services/*/Dockerfile` |
| Migration-bearing services | 40 | `services/*/migrations` |
| Tools / SDK (not container matrix) | tools/*, packages/sdk/go | Covered by go test in their modules when run separately; acceptance focuses on services |

## Acceptance matrix (this environment)

| Gate | Result | Evidence |
|------|--------|----------|
| 46/46 test | **PASS** (prior) | Prompt 51/52 local `PASS=46`; not re-run full matrix this pass |
| 46/46 build | **PASS** (prior) | Prompt 51/52 local `PASS=46` |
| go mod verify | **PASS** (prior sample) | Prompt 51 sample; CI job now runs verify for all |
| race | **BLOCKED** | No gcc locally; CI job ready: `go-race-all` |
| Docker 46/46 | **BLOCKED** | No Docker locally; CI job ready: `docker-build-all` |
| Compose config | **BLOCKED** | No Docker; CI: `docker compose … config -q` |
| Compose startup | **BLOCKED** | No Docker; CI migration/startup jobs |
| PostgreSQL migration | **BLOCKED** | Script applies all migration dirs dynamically on CI |
| Redis | **BLOCKED** | CI: PING + SET/GET |
| Kafka | **BLOCKED** | CI: broker ready + topic `nexora.ci.smoke` |
| service startup | **BLOCKED** | CI: `service-startup-smoke.sh` (core 12 services) |
| E2E | **BLOCKED** | No full API workflow in CI yet; startup/health is partial |
| CI workflow | **PASS** (static) | YAML + scripts reviewed; **execution not run** |
| security sanity | **PASS** (delta) | `permissions: contents: read`; no secrets added |

## Blockers

| WHAT | WHY | LIMITATION | CI CAN VERIFY? | CHANGED |
|------|-----|------------|----------------|---------|
| Actual race/Docker/compose/migration/startup | No Docker, no CGO toolchain, no `gh` | Local Windows agent | Yes after push/`workflow_dispatch` | Hardened workflow + scripts |
| Full E2E commerce path | Not implemented as automated CI script | Scope | Partial via startup smoke | Documented BLOCKED |

## Identity SessionCache

Unchanged from Prompt 52: **not injected**; `SessionRepository` remains SoT.
