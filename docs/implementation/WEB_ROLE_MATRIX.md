# Web role matrix (Prompt 66)

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
| admin | identity seed | tenant | `admin_web` | `/dashboard` `/orders` … | admin:* | `bff-admin` + mocks | full ops UI | **Partial** (many mocks) |
| super_admin | identity seed | global | `super_admin_web` | `/tenants` `/flags` … | platform:* | mock/platform | tenant governance | **Partial** (mock-heavy) |
| service_account | identity seed | machine | — | — | — | internal | automation | **N/A web** |

## Authentication

| App | Method | Endpoint |
|-----|--------|----------|
| customer-web | OTP via BFF | `POST /v1/customer/auth/otp/*` |
| All staff role apps | OTP via identity | `POST /v1/identity/auth/otp/*` |
| admin_web / super_admin_web | Demo login (pending identity wire) | local store |

## Tenant isolation

- All API clients send `X-Tenant-Id` from `NEXT_PUBLIC_TENANT_ID`.
- Automated API tests: `qa/playwright/tests/tenant.isolation.spec.ts`.
- Backend authorization is authoritative; UI `RouteGuard` is UX-only.

## Known gaps

1. `admin_web` / `super_admin_web` — majority mock APIs; `bff-admin` has 6 live routes.
2. Public staging URLs — blocked without `VERCEL_TOKEN` + public BFF host.
3. Multi-role E2E chain — requires persistent staging DB + full service mesh.
4. Realtime SSE — customer tracking uses polling until gateway WebSocket exists.

---

## Prompt 66 acceptance (STATE B)

Public staging **BLOCKED** without `VERCEL_TOKEN` + public BFF URL. Customer + staff OTP integrated; admin/super_admin remain demo. CI web-e2e via `scripts/ci/web-e2e.sh`. See commit after Prompt 66 push for final hash.
