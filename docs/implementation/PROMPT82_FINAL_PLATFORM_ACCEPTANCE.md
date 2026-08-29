# Final web platform acceptance — audit, fixes and fresh verification

**HEAD at verification:** `eb5ffaa`
**Date:** 2026-08-29
**Environment:** Codespace `ominous-sniffle-v66jrrr567vcwpvj`, `nexora-phone-staging` Docker network (18 services), `customer-web` dev server on `127.0.0.1:3000`
**Order under test:** `eeddf39a-69cc-4edb-8f7c-48802d38f822` (fresh, created during this pass)

## What this pass fixed

### Login roles were never enforced (`48db732`)

`verifyOtp` accepted an `expectedRoles` argument and discarded it, so an identity
principal could sign in to any console. The admin console additionally granted a
fallback `viewer` session and asserted `mfaVerified: true` without any MFA step, and
the six staff dashboards gated on session presence rather than role. Roles are now
asserted at verification (`RoleNotAllowedError`) and every staff dashboard is wrapped
in a role-scoped `RouteGuard`.

### Fabricated prices (`48db732`)

The product page fetched nothing: it invented a name from the URL and added a
hardcoded 15.00 line to the cart. Home substituted 10.00 whenever the feed carried no
price. The product page now reads the real catalog item and says plainly that a price
is calculated at checkout when none is published.

### Storefront died with a supplementary rail (`48db732`)

`bff-customer` returned `502` for every signed-in `/v1/customer/home` request when the
recommendation upstream was unavailable. Rails are now best-effort; catalog and
serviceability failures still surface. Covered by `TestHomeSurvivesRecsOutage` and
`TestHomeFailsWhenCatalogDown`.

### Tailwind was never loaded in seven of nine apps (`1e37549`)

Tailwind v4 was installed and wired into PostCSS everywhere, but only the two admin
consoles imported it from their stylesheet. The customer app and the courier,
warehouse, supplier, finance, support and operations consoles therefore shipped markup
full of utility classes with no matching rules — flex containers fell back to inline
layout, padding vanished, and `min-height` on inline anchors was ignored. This is why
mobile touch targets measured 17px however tall the markup asked to be. Each
stylesheet now imports `tailwindcss` before the design tokens and declares the shared
`@nexora/web-core` sources.

### Other application fixes (`48db732`, `673db14`, `eb5ffaa`)

- `401` now clears the session and returns to sign-in; non-JSON and 5xx upstream bodies are never rendered.
- Checkout reuses one cart id and one idempotency key per attempt, so a double click cannot place two orders.
- Raw `JSON.stringify` dumps on the order, order detail and checkout screens were replaced with real summaries.
- The account screen asked the browser for a real position instead of writing one fixed coordinate pair for every customer.
- Turbopack could not resolve `@nexora/web-core` because the inferred root stopped at the app directory; all nine `next.config.ts` files now point it at the workspace root.
- `scripts/ci/web-static.sh` runs the typecheck and vitest suites that existed but were never executed.
- Navigation, back links, cart actions and the home product card meet the 44px touch target floor.

## Fresh verification at `eb5ffaa`

### `prompt80-evidence.py`

| Check | Result |
|-------|--------|
| Customer OTP login via same-origin `/v1/customer/auth/otp/*` | `200`, roles `[customer]` |
| Catalog, cart add, checkout preview, place | `200` / `201` / `200` / `201` |
| RBAC denials (7 cross-role calls) | all `403` |
| Legitimate role APIs (7 roles) | all `200` |
| Wrong-tenant order read | `404` (right tenant `200`) |
| Wrong-tenant cart write / warehouse pick | `404` / `404` |
| Unauthenticated SSE | `401` |
| Own-order SSE with ticket | `200`, `: connected` |
| Cross-topic SSE with own ticket | `403` |
| Wrong-tenant ticket mint / other customer's order and ticket | `404` / `404` / `404` |

Verdict `PROMPT80_EVIDENCE_PRE_SSE_OK`.

### `prompt80-browser-sse.mjs` (Chromium, 390×844 and 393×852)

- Warehouse `pick` (`200`) delivered a live event to the open browser: `Event: picking`, `SSE connected; live event received.`, tracking UI status `picking` — no reload, no polling fallback.
- Horizontal overflow `0` on login, home, search, cart, checkout, orders, order detail and track, at both viewports.
- Touch targets under 44px: none.
- Product page shows the real catalog item with no fabricated price.
- Unauthenticated deep link to `/orders/<id>/track` redirects to `/login` and leaks no order data.

Verdict `PROMPT80_BROWSER_OK`.

### `prompt80-after-sse.py`

| Transition | HTTP | Customer-visible status |
|------------|------|-------------------------|
| warehouse pack | 200 | `ready_for_dispatch` |
| warehouse ready | 200 | `ready_for_dispatch` |
| courier accept | 200 | `courier_assigned` |
| courier enroute | 200 | `out_for_delivery` |
| courier complete | 200 | `completed` |
| BFF track final | 200 | `completed` |
| tracking-service timeline direct | 200 | `items` present |

Verdict `PROMPT80_AFTER_SSE OK`.

### Builds and tests

- `next build` for all nine web apps: pass.
- `go test ./...` across every module in the Go workspace: pass.
- `customer-web` `tsc --noEmit` and `vitest run` (8 tests): pass.

## Environment notes

- Catalog, cart and order services keep state in memory, so a Codespace reboot empties the catalog. Reseed with `scripts/staging/prompt80-seed-catalog.sh` before running the evidence scripts.
- Turbopack's dev cache can survive a `git reset` and keep serving the previous module graph. When verifying a CSS or layout change, stop the dev server, delete `apps/customer-web/.next` and start it again.
- Windows→Codespace copies arrive with CRLF; run `sed -i "s/\r$//"` on shell scripts before executing them.
- The staging URLs are Codespace port forwards behind GitHub authentication. There is no permanent public deployment.
