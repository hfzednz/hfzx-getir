# Web role matrix (Prompt 66–67)

Evidence-backed mapping from identity seed, BFFs, and service APIs.

| Role | Source | Tenant scope | Web app | Routes | Permissions (seed) | Backend APIs | Core workflows | Status |
|------|--------|--------------|---------|--------|-------------------|--------------|----------------|--------|
| customer | `019_seed_roles_permissions.sql` | tenant | `customer-web` | `/login` `/home` `/cart` `/checkout` `/orders/*` | orders:read/write, identity:self | `bff-customer` | OTP, browse, cart, checkout, track | **Integrated** |
| courier | identity seed | tenant | `courier-web` | `/login` `/dashboard` | identity:self | `bff-courier` duty, offers | online/offline, accept offer | **Integrated** (stub BFF data) |
| picker | identity seed | tenant/warehouse | `warehouse-web` | `/login` `/dashboard` | warehouse:pick | `bff-warehouse` pick/pack/ready | pick → pack → ready | **Integrated** |
| packer | identity seed | tenant/warehouse | `warehouse-web` | same | warehouse:pack | same | same | **Integrated** |
| dispatcher | identity seed | tenant/warehouse | `warehouse-web` | same | warehouse:dispatch | same | ready dispatch | **Integrated** |
| supplier | identity seed | tenant | `supplier-web` | `/login` `/dashboard` | partner APIs | `supplier-service` | suppliers, POs | **Integrated** |
| seller/merchant | `supplier-service` sellers | tenant/seller | `supplier-web` | same | listing/catalog APIs | `supplier-service` | catalog, listings | **Partial** |
| partner | identity seed | tenant | `supplier-web` | same | supplier APIs | `supplier-service` | shared portal | **Partial** |
| city_ops | identity seed | tenant/city | `operations-web` | `/login` `/dashboard` | admin:read (UI) | `bff-admin` dashboard | ops overview | **Integrated** (admin BFF minimal) |
| support_agent | identity seed | tenant | `support-web` | `/login` `/dashboard` | support:read/write | `bff-admin` orders | lookup orders | **Integrated** (partial) |
| finance_analyst | identity seed | tenant | `finance-web` | `/login` `/dashboard` | finance:read | `finance-ledger-service` | journals, ledger | **Integrated** |
| admin | identity seed | tenant | `admin_web` | `/dashboard` `/orders` … | admin:* | `bff-admin` (5 routes) | dashboard, orders | **Real OTP** — extended UI needs backend |
| super_admin | identity seed | global | `super_admin_web` | `/tenants` `/flags` … | platform:* | `platform-ops-service` | stats, deployments | **Real OTP** — most screens need backend |
| service_account | identity seed | machine | — | — | — | internal | automation | **N/A web** |

## Authentication

| App | Method | Endpoint |
|-----|--------|----------|
| customer-web | OTP via BFF | `POST /v1/customer/auth/otp/*` |
| All staff role apps | OTP via identity | `POST /v1/identity/auth/otp/*` |
| admin_web | OTP via identity | Real OTP |
| super_admin_web | OTP via identity | Real OTP |

## Tenant isolation

- All API clients send `X-Tenant-Id` from `NEXT_PUBLIC_TENANT_ID`.
- Automated API tests: `qa/playwright/tests/tenant.isolation.spec.ts`.
- Backend authorization is authoritative; UI `RouteGuard` is UX-only.

## Known gaps

1. Extended admin/super-admin UI modules lack backend routes (mock fallback disabled by default).
2. Public staging URLs — blocked without `VERCEL_TOKEN` + public BFF host.
3. Full browser multi-role E2E — API journey only; warehouse needs task ID env.
4. Realtime SSE client wired; publishers from order-service not fully integrated.

---

## Prompt 66 acceptance (STATE B)

Public staging **BLOCKED** without `VERCEL_TOKEN` + public BFF URL. Customer + staff OTP integrated; admin/super_admin remain demo. CI web-e2e via `scripts/ci/web-e2e.sh`. See commit after Prompt 66 push for final hash.
