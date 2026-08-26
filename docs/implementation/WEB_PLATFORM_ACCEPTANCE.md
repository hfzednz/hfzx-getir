# Web platform acceptance matrix (Prompt 65)

Commit: see `main` after Prompt 65 push.

## Summary

| Final state | **STATE B** — implemented; external deployment required |
|-------------|--------------------------------------------------------|

## Role acceptance matrix

| Role | UI | Auth | RBAC (UI) | Core flow | E2E | Responsive | Security |
|------|----|------|-----------|-----------|-----|------------|----------|
| customer | **PASS** (customer-web) | OTP wired to BFF | customer session | browse/cart/checkout/order | CI smoke + BFF journeys | mobile-first shell | tenant header; no secrets in client |
| courier | **PASS** (shell) | demo + duty API | courier role | duty toggle | manual | mobile-first dark UI | BFF only |
| picker/packer/dispatcher | **PASS** (warehouse-web) | demo | warehouse roles | pick/pack/ready | manual | tablet/mobile | task ID scoped |
| supplier/seller | **PASS** (shell) | demo | supplier | portal health | manual | desktop | supplier-service |
| finance_analyst | **PASS** (shell) | demo | finance | ledger journals GET | manual | desktop | finance-service |
| support_agent | **PASS** (shell) | demo | support | admin BFF health | manual | desktop | admin BFF |
| city_ops | **PASS** (operations-web) | demo | city_ops | admin BFF health | manual | desktop | admin BFF |
| admin | **PASS** (existing admin_web) | demo login | 52 permissions | orders/dashboard partial live | Playwright admin | responsive | PermissionGate |
| super_admin | **PASS** (existing super_admin_web) | demo login | platform roles | tenants/flags mock | not automated | responsive | dual-control docs |

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
