# 13 — Navigation

## Principles

1. Users always know **where they are** and **how to go back**
2. Customer: progressive commitment — browse light, checkout focused
3. Ops: task-first — minimize chrome during active pick/delivery
4. Admin: deep-linkable entity URLs + city context sticky
5. Animations clarify hierarchy — never block input

---

## Shells by surface

### Customer — Bottom Navigation

Items (max 5): Home · Search · Cart/Orders · Account (exact IA finalized per app prompt; Cart may badge into Orders hub).

- Active: filled icon + `nav.itemActive`
- Inactive: outlined + `nav.itemDefault`
- Height: 56 + safe area
- Hide on immersive flows: checkout, auth OTP, media fullscreen

### Nested navigation

- Stack per tab (indexed stack) — tab switch restores state
- PDP / category pushed on Home or Search stack
- Modal routes for filters, address picker, auth

### Courier

Bottom nav lighter: Tasks · Earnings · Account; map takes primary chrome during active delivery (collapse nav).

### Warehouse

Often **no** bottom nav during shift mode — station top bar + task queue. Break-mode may show limited nav.

### Tablet — Navigation Rail

72–80px rail; icons + short labels; same destinations as bottom nav.

### Desktop — Sidebar

- Admin / Super Admin: `NxSidebar` 240–264px
- Sections: Ops, Catalog, Users, Finance, Config…
- City switcher pinned top
- Collapse to icon rail at `md` width
- Super Admin: platform section separated visually (border/overline)

---

## Search navigation

- Search tab or top-bar entry → full search route
- Suggestions overlay then results
- Back clears query stack sanely (one back from results → suggestions/home)

## Deep linking & universal links

| Pattern | Example |
|---------|---------|
| Product | `nexora://product/{id}` / `https://nexora.app/p/{id}` |
| Category | `/c/{slug}` |
| Order | `/orders/{id}` |
| Promo | `/promo/{code}` |
| Store (ops) | `/ops/stores/{id}` |

Rules:

- Auth-gated links buffer intended path post-login
- Expired promo → friendly fallback + home
- City mismatch → prompt switch or explain unavailability

---

## Navigation animations

| Transition | Spec |
|------------|------|
| Push (forward) | Shared axis X 8–12% + fade; `short` 260ms; emphasized |
| Pop | Reverse |
| Tab switch | Fade only `fast` 120ms — no slide war |
| Modal / sheet | Vertical slide + scrim fade `medium` 320ms |
| Checkout enter | Vertical full-screen `medium` |
| Hero (image → PDP) | Optional shared element; disable if reduced motion |

GoRouter / web router must map to these tokens.
