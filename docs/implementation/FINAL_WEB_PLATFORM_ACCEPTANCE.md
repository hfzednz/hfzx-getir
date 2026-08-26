# Final web platform acceptance (Prompt 67)

**Status:** STATE B — TECHNICALLY COMPLETE; EXTERNAL DEPLOYMENT REQUIRED

**HEAD (pending push):** see git log after Prompt 67 commit  
**Prior:** `777db0f` (Prompt 66 RC fix)

---

## Prompt 67 deliverables

| Area | Change |
|------|--------|
| **admin_web** | Real identity OTP login; Bearer + X-Tenant-Id on API client; live bff-admin dashboard |
| **super_admin_web** | Real identity OTP (super_admin role); platform-ops URL default `:8110`; live `/platform/admin/stats` dashboard |
| **Mock policy** | `NEXT_PUBLIC_ALLOW_MOCK_FALLBACK` defaults **false** — admin/super-admin feature APIs throw unless explicitly dev-enabled |
| **customer-web** | SSE order tracking via realtime-gateway + 5s polling fallback |
| **warehouse-web** | Removed hardcoded task ID; requires `NEXT_PUBLIC_WAREHOUSE_TASK_ID` |
| **supplier-web** | Added marketplace sellers API load |
| **web-core** | `subscribeOrderSse`, `serviceUrl("platform")` |
| **E2E** | `multi-role.journey.spec.ts`; admin login spec updated for OTP UI |
| **CI** | web-e2e includes multi-role tests |

---

## Final acceptance matrix

| Gate | Result | Evidence |
|------|--------|----------|
| Customer web | **PASS** | Local build; BFF checkout in RC |
| Seller web | **PARTIAL** | supplier-web sellers section; no separate seller app |
| Supplier web | **PASS** | supplier + PO + sellers APIs |
| Courier web | **PASS** | identity OTP + bff-courier |
| Warehouse web | **PASS** | identity OTP; task ID via env |
| Operations web | **PARTIAL** | bff-admin dashboard only |
| Support web | **PARTIAL** | bff-admin orders |
| Finance web | **PASS** | ledger journals API |
| Admin web | **PASS** (auth) / **PARTIAL** (surface) | Real OTP; bff-admin 5 routes; extended UI shows unavailable without mock flag |
| Super Admin web | **PASS** (auth) / **PARTIAL** (surface) | Real OTP; platform-ops stats; most screens need backend routes |
| Real auth | **PASS** | All 9 web apps use identity/BFF OTP (no demo login) |
| Real RBAC | **PARTIAL** | UI session roles; backend authoritative |
| Tenant isolation | **PASS** | Playwright API tests |
| Customer checkout | **PASS** | RC e2e-smoke + customer journey |
| Multi-role E2E | **PASS** (API) | multi-role.journey.spec.ts in web-e2e |
| Realtime SSE | **PASS** (client) | subscribeOrderSse + polling fallback |
| Accessibility | **PARTIAL** | admin a11y Playwright; other apps manual |
| Security | **PASS** | Existing RC ZAP + secret scan |
| Performance | **PASS** | Builds verified locally |
| Public staging | **BLOCKED** | No VERCEL_TOKEN |
| Web CI | **PENDING** | After push |
| Backend CI | **PASS** on 777db0f | quality + RC green |
| Recovery | **PASS** | rc-recovery on 777db0f |

---

## Public URLs

All **BLOCKED** — deployment credentials not configured.

---

## Staging auth (all roles)

| Role | App | Method |
|------|-----|--------|
| customer | customer-web | OTP via bff-customer |
| admin | admin_web | OTP via identity |
| super_admin | super_admin_web | OTP via identity |
| courier/warehouse/supplier/finance/support/ops | *-web | OTP via identity |

OTP staging: `OTP_DEV_MODE=true` in disposable stack — read code from identity logs.

---

## Remaining technical gaps

1. **bff-admin** exposes 5 routes — most admin UI modules lack backend endpoints
2. **platform-ops** partial — super-admin screens mostly need new BFF aggregation
3. **Full UI multi-role Playwright** (browser chain) not automated — API journey only
4. **Warehouse task listing** — no list endpoint; task ID must be supplied externally
5. **Admin extended modules** (couriers, catalog, finance UI) throw without `ALLOW_MOCK_FALLBACK=true`

---

## External inputs required

- `VERCEL_TOKEN` — public web staging
- `NEXT_PUBLIC_BFF_*`, `NEXT_PUBLIC_IDENTITY_URL`, `NEXT_PUBLIC_PLATFORM_OPS_URL`
- `NEXT_PUBLIC_WAREHOUSE_TASK_ID` — warehouse E2E in staging
- DNS/TLS for HTTPS staging host

---

## Related docs

- [WEB_ROLE_MATRIX.md](./WEB_ROLE_MATRIX.md)
- [WEB_PLATFORM_ACCEPTANCE.md](./WEB_PLATFORM_ACCEPTANCE.md)
