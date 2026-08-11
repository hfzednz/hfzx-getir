# NEXORA Monorepo Structure (Prompt-44)

Architecture and service names are **frozen**. This document is the on-disk assembly map.

Do **not** rename folders to match generic examples in prompts. Use the **canonical paths** below.

## Top-level

| Path | Role |
|------|------|
| `apps/` | Customer / courier / warehouse Flutter + admin / super-admin web |
| `services/` | Go microservices + BFFs + gateways |
| `packages/` | Flutter design/core, web UI, SDKs, innovation stubs |
| `infra/` | Terraform, Helm, Kustomize, Argo CD, Docker, observability |
| `ops/` | Runbooks, playbooks, release, SLO, production ops |
| `docs/` | Constitution, design system, production, API, guides |
| `qa/` | Quality suites (Playwright, k6, chaos, hyperscale) |
| `tools/` | Certifiers / validators (`integration-cert`, `prod-validate`, …) |
| `store/` | App Store / Play ASO copy |
| `scripts/` | Bootstrap / verify / doctor |
| `ADR/` | Architecture decision records |
| `.github/` | CI/CD workflows + CODEOWNERS |

## Alias map (prompt examples → canonical)

| Example name | Canonical path |
|--------------|----------------|
| `customer-mobile` | `apps/mobile_customer` |
| `courier-mobile` | `apps/mobile_courier` |
| `warehouse-mobile` | `apps/mobile_warehouse` |
| `admin-dashboard` | `apps/admin_web` |
| `super-admin` | `apps/super_admin_web` |
| `auth-service` | `services/identity-service` |
| `customer-service` | `services/customer-profile-service` |
| `ai-service` | `services/ai-platform-service` |
| `analytics-service` | `services/data-platform-service` |
| `infrastructure/` | `infra/` |
| `github/` | `.github/` |
| `design-system` (code) | `packages/flutter/nexora_design` + `packages/web/ui` + `docs/design-system/` |
| `ui-components` | `packages/web/ui` (`@nexora/ui`) |
| `localization` | per-app l10n + `services/global-service` |
| `sdk` | `packages/sdk/{go,flutter,nodejs,python}` |
| `runbooks` / `playbooks` | `ops/runbooks`, `ops/playbooks` |
| `testing` | `qa/` + per-module `*_test.go` / Flutter tests |

## Workspace manifests

| Ecosystem | File |
|-----------|------|
| Go | `go.work` (all services + tools + SDK) |
| Flutter | `melos.yaml` |
| Node / web | `pnpm-workspace.yaml` + root `package.json` |
| Make | `Makefile` |

## Clone-and-run

```bash
git clone <repo> && cd hfzx_Getir_
# Windows:
pwsh -File scripts/bootstrap.ps1
# Unix:
bash scripts/bootstrap.sh
make verify
```

## Known nesting debt (do not use)

Accidental nested trees may exist under `apps/apps`, `apps/packages`, or `apps/mobile_customer/apps/…`.  
**Canonical apps live directly under `apps/{mobile_*,admin_web,super_admin_web}`.** Nested copies are non-canonical; do not point CI at them.

## Service registry

Ports and ownership: `docs/launch/service-registry.yaml`.
