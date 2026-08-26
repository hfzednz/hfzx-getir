# Web platform acceptance matrix (Prompt 65–66)

Commit: Prompt 65 `cdda568`; Prompt 66 changes pending push (real OTP, web E2E).

## Summary

| Final state | **STATE B** — technically complete; external deployment required |
|-------------|-------------------------------------------------------------------|

Prompt 66: real identity/BFF OTP on role apps, shared `@nexora/web-core` auth, tenant isolation Playwright tests, `scripts/ci/web-e2e.sh`, `cd-web-staging.yml`. See `WEB_ROLE_MATRIX.md`.

## Role acceptance matrix

| Role | UI | Auth | RBAC (UI) | Core flow | E2E | Responsive | Security |
|------|----|------|-----------|-----------|-----|------------|----------|
| customer | **PASS** (customer-web) | OTP wired to BFF | customer session | browse/cart/checkout/order | CI smoke + BFF journeys | mobile-first shell | tenant header; no secrets in client |
| courier | **PASS** (courier-web) | OTP via identity | courier session | duty toggle + offers | CI web-e2e | mobile-first dark UI | BFF only |
| picker/packer/dispatcher | **PASS** (warehouse-web) | OTP via identity | warehouse session | pick/pack/ready | CI web-e2e | tablet/mobile | task ID scoped |
| supplier/seller | **PASS** (supplier-web) | OTP via identity | supplier session | suppliers + POs | CI web-e2e | desktop | supplier-service |
| finance_analyst | **PASS** (finance-web) | OTP via identity | finance session | ledger journals GET | CI web-e2e | desktop | finance-service |
| support_agent | **PASS** (support-web) | OTP via identity | support session | bff-admin orders | CI web-e2e | desktop | admin BFF (minimal) |
| city_ops | **PASS** (operations-web) | OTP via identity | ops session | bff-admin dashboard | CI web-e2e | desktop | admin BFF (minimal) |
| admin | **PARTIAL** (admin_web) | **OTP via identity** | 52 permissions UI | orders/dashboard live | Playwright admin | responsive | PermissionGate |
| super_admin | **PARTIAL** (super_admin_web) | **OTP via identity** | platform roles UI | platform stats live | not automated | responsive | dual-control docs |

## Customer web route map

| Route | Purpose |
|-------|---------|
| `/login` | OTP |
| `/home` | Catalog home |
| `/search` | Search |
| `/product/[id]` | Product detail |
| `/cart` | Cart |
| `/checkout` | Preview + place |
| `/orders` | History |
| `/orders/[id]` | Detail |
| `/orders/[id]/track` | Tracking |
| `/account` | Profile + location |

## Operational web route map

| App | Base (dev) | Routes |
|-----|------------|--------|
| courier-web | `:3001` | `/login`, `/dashboard` |
| warehouse-web | `:3002` | `/login`, `/dashboard` |
| supplier-web | `:3003` | `/login`, `/dashboard` |
| finance-web | `:3004` | `/login`, `/dashboard` |
| support-web | `:3005` | `/login`, `/dashboard` |
| operations-web | `:3006` | `/login`, `/dashboard` |
| admin_web | `:3100` | see `nav-config.ts` |
| super_admin_web | dev default | see `nav-config.ts` |

## E2E results

| Suite | Result | Evidence |
|-------|--------|----------|
| BFF customer journey | **PASS** | `ci-acceptance` e2e-smoke |
| Admin Playwright | **PASS** | `ci-release-candidate` rc-ui-a11y |
| Customer web build smoke | **PASS** (CI) | `ci-web-quality` customer-web-e2e job |
| Multi-role chain | **NOT EXECUTED** | requires deployed full stack |
| Tenant isolation automated | **NOT EXECUTED** | manual test plan in QA strategy |

## External requirements

1. `VERCEL_TOKEN` or hosting platform credentials per app
2. Public BFF URLs (`NEXT_PUBLIC_BFF_*`)
3. OIDC / identity integration for production auth (replace demo login on ops apps)
4. CORS configuration on BFFs for web origins
5. Realtime SSE/WebSocket wiring for live tracking

## Regression

Existing Quality / Acceptance / RC / Recovery pipelines — **not modified** (only additive `ci-web-quality.yml`).
