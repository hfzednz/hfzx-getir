# Full web platform — NEXORA (Prompt 65)

Web-first channel added on top of the existing backend. Flutter mobile apps unchanged.

## Final state

**STATE B — FULL WEB PLATFORM — IMPLEMENTED, EXTERNAL DEPLOYMENT REQUIRED**

Web applications are implemented and build locally/CI. Public URLs require Vercel/hosting credentials and a reachable staging/production BFF. No fabricated deployment URLs.

## Role discovery matrix

| Role | Identity seed | Web app | Port (dev) | API surface |
|------|---------------|---------|------------|-------------|
| customer | `customer` | `apps/customer-web` | 3000 | `bff-customer` |
| courier | `courier` | `apps/courier-web` | 3001 | `bff-courier` |
| picker | `picker` | `apps/warehouse-web` | 3002 | `bff-warehouse` |
| packer | `packer` | `apps/warehouse-web` | 3002 | `bff-warehouse` |
| dispatcher | `dispatcher` | `apps/warehouse-web` | 3002 | `bff-warehouse` |
| supplier | `supplier` | `apps/supplier-web` | 3003 | `supplier-service` |
| seller/merchant | via `supplier-service` sellers API | `apps/supplier-web` `/dashboard` | 3003 | `supplier-service` |
| partner | `partner` | `apps/supplier-web` (shared) | 3003 | `supplier-service` |
| finance_analyst | `finance_analyst` | `apps/finance-web` | 3004 | `finance-ledger-service` |
| support_agent | `support_agent` | `apps/support-web` | 3005 | `bff-admin` + admin modules |
| city_ops | `city_ops` | `apps/operations-web` | 3006 | `bff-admin` |
| admin | `admin` | `apps/admin_web` | 3100 | `bff-admin` |
| super_admin | `super_admin` | `apps/super_admin_web` | 3200* | platform APIs (mock/BFF partial) |

\* super_admin_web default dev port from package scripts (use `npm run dev`).

### Mobile-only roles (no dedicated web in this pass)

| Role | Mobile app |
|------|------------|
| picker/packer/dispatcher (extended) | `apps/mobile_warehouse` |
| courier (extended) | `apps/mobile_courier` |
| customer (extended) | `apps/mobile_customer` |

Warehouse web covers pick/pack/ready via `bff-warehouse`. Extended warehouse manager roles remain in Flutter.

## Architecture

```
packages/web/
  ui/          @nexora/ui — design system
  core/        @nexora/web-core — API client, auth store, RBAC helpers, order states

apps/
  customer-web/      Getir-style responsive storefront
  courier-web/       Mobile-first delivery console
  warehouse-web/     Pick / pack / ready
  supplier-web/      Supplier + seller portal shell
  finance-web/       Ledger / settlement shell
  support-web/       Support console shell
  operations-web/    City ops shell
  admin_web/         Existing tenant admin (22+ modules)
  super_admin_web/   Existing platform super-admin
```

## Customer web flows

| Flow | Route | BFF |
|------|-------|-----|
| OTP login | `/login` | `POST /v1/customer/auth/otp/*` |
| Home / browse | `/home` | `GET /v1/customer/home` |
| Search | `/search?q=` | `GET /v1/customer/home?q=` |
| Product | `/product/[id]` | local + cart |
| Cart | `/cart` | local state + `POST /v1/customer/cart/items` |
| Checkout | `/checkout` | preview + place |
| Orders | `/orders`, `/orders/[id]` | `GET /v1/customer/orders/{id}` |
| Tracking | `/orders/[id]/track` | polling `GET .../track` (SSE when realtime wired) |
| Account | `/account` | session + location |

Order states in UI use canonical `order-service` states via `@nexora/web-core`.

## Shared packages

| Package | Purpose |
|---------|---------|
| `@nexora/web-core` | `createApiClient`, `createSessionStore`, `bffUrl`, `serviceUrl`, `RoleGate`, `ORDER_STATES` |
| `@nexora/ui` | Buttons, grids, shells, tokens |

## Environment variables

| Variable | Used by |
|----------|---------|
| `NEXT_PUBLIC_BFF_CUSTOMER_URL` | customer-web |
| `NEXT_PUBLIC_BFF_ADMIN_URL` | admin, support, operations |
| `NEXT_PUBLIC_BFF_COURIER_URL` | courier-web |
| `NEXT_PUBLIC_BFF_WAREHOUSE_URL` | warehouse-web |
| `NEXT_PUBLIC_SUPPLIER_URL` | supplier-web |
| `NEXT_PUBLIC_FINANCE_URL` | finance-web |
| `NEXT_PUBLIC_TENANT_ID` | all |

## CI

| Workflow | Purpose |
|----------|---------|
| `ci-web-quality.yml` | `scripts/ci/web-static.sh` — build all web apps |
| Existing `ci-acceptance`, `ci-release-candidate`, `ci-quality` | unchanged |

## Deployment

1. Deploy backend (see `docs/implementation/PHONE_TEST_ENVIRONMENT.md`).
2. Deploy each Next.js app to Vercel/Node host with env vars above.
3. Preview deployments via `workflow_dispatch` (add `cd-web-preview.yml` when `VERCEL_TOKEN` available).

## Known limitations

1. **No public URLs** without hosting credentials.
2. **admin_web / super_admin_web** — many modules still mock data; BFF-admin has 6 live routes vs large UI surface.
3. **bff-courier / bff-warehouse** — stub in-memory journeys; web UIs call real routes.
4. **Realtime** — gateway exposes SSE only; customer tracking uses polling.
5. **Seller** — no separate app; supplier-service seller APIs exposed via supplier-web shell.
6. **Settlement UI** — finance-web shell; full settlement batches via `settlement-service` API (integrate next).
7. **RBAC** — frontend gates are UX-only; backend authorization remains authoritative.
8. **Three role systems** — identity seed vs admin_web vs super_admin_web permissions not fully unified.

## Related

- `docs/implementation/WEB_PLATFORM_ACCEPTANCE.md` — acceptance matrix
- `docs/implementation/PHONE_TEST_ENVIRONMENT.md` — backend + APK
- `docs/design-system/00-INDEX.md` — visual constitution
