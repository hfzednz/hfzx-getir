# Prompt 80 — Fresh order, RBAC, tenant isolation, live SSE

**HEAD:** `8650205` (+ local fix to `scripts/staging/prompt80-browser-sse.mjs`)
**Date:** 2026-08-29
**Environment:** Codespace `ominous-sniffle-v66jrrr567vcwpvj`, `nexora-phone-staging` Docker network (18 services), `customer-web` dev server on `127.0.0.1:3000`
**STATE:** PASS
**Order under test:** `81e57353-7ea4-4057-8450-6f17c338bced`

## Environment bring-up

- Codespace was shut down; all 18 `nexora-staging-*` containers restarted from existing `phone-staging` images. Health `200` on `8081, 8091, 8110, 8111, 8112, 8113, 8114, 8115, 8117`.
- Catalog/cart/order services are in-memory, so a reboot leaves the catalog empty and `prompt80-evidence.py` fails at `FAIL no catalog`. Added `scripts/staging/prompt80-seed-catalog.sh` (extracted from `prompt76-codespace-apply.sh`) to reseed product `fresh-milk` / variant `MILK-1L` and reindex search.
- Windows→Codespace file copies arrive with CRLF; scripts need `sed -i "s/\r$//"` before running.

## `prompt80-evidence.py`

| Check | Result |
|-------|--------|
| Customer OTP login via same-origin `/v1/customer/auth/otp/*` | 200, JWT roles `[customer]` |
| Catalog via `/v1/customer/home` | 200, 1 product |
| Cart add / checkout preview / place | 201 / 200 / 201 |
| Order read + BFF track + tracking-service direct | 200 / 200 / 200, status `warehouse_assigned` |
| RBAC denials (7 cross-role calls) | all `403 forbidden` |
| Legitimate role APIs (7 roles) | all `200` |
| Wrong-tenant order read | `404` (right tenant `200`) |
| Unauthenticated SSE | `401` |
| Own-order SSE with ticket | `200`, stream opens with `: connected` |
| Cross-topic SSE with own ticket | `403` |
| Wrong-tenant ticket mint | `404` |
| Other customer's order read / ticket mint | `404` / `404` |

Verdict: `PROMPT80_EVIDENCE_PRE_SSE_OK`

## `prompt80-browser-sse.mjs` (Chromium, iPhone viewport 390×844)

- Fixed the script: it was an ESM file using `require` with a hardcoded `/workspaces/hfzx-getir/...` Playwright path, so it could never run. Now resolves `@playwright/test` through `qa/playwright` via `createRequire`.
- Login → `/home` → `/search` → `/orders/<id>/track`, all with `scrollWidth - clientWidth = 0` (no mobile overflow).
- Track page minted a realtime ticket (`200`) and opened `GET /v1/realtime/sse?topic=order:<id>&ticket=…` → `200`; UI showed `SSE connected (waiting for event).`
- Warehouse `pick` (`200`) pushed a live event: UI moved to `Event: picking`, `SSE connected; live event received.`, `track-status` = `picking` — no page reload, no polling fallback.

Verdict: `PROMPT80_BROWSER_OK`

## `prompt80-after-sse.py`

| Transition | HTTP | Customer-visible status |
|------------|------|-------------------------|
| warehouse pack | 200 | `ready_for_dispatch` |
| warehouse ready | 200 | `ready_for_dispatch` |
| courier accept | 200 | `courier_assigned` |
| courier enroute | 200 | `out_for_delivery` |
| courier complete | 200 | `completed` |
| BFF track final | 200 | `completed` |
| tracking-service timeline direct | 200 | `items` present |

Verdict: `PROMPT80_AFTER_SSE OK`

## Known noise (not a product defect)

The browser run logged two `502 Bad Gateway` console errors after `/home` and `/search`. Neither the Next dev server request log nor `bff-customer` logs contain any `502`; a re-probe showed the failing requests are Next dev RSC prefetches (`/login?_rsc=…`) during Fast Refresh. All product endpoints on those pages returned `200`.
