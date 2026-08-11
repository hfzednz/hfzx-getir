# 21 — Responsive & Adaptive System

## Breakpoints

| Token | Min width | Class |
|-------|-----------|-------|
| `xs` | 0 | Mobile |
| `sm` | 600 | Large phone / small tablet |
| `md` | 905 | Tablet |
| `lg` | 1240 | Desktop |
| `xl` | 1440 | Large desktop |
| `xxl` | 1920 | Ultra wide |

Helpers: `NxBreakpointUtils` / CSS media queries from tokens.

---

## Layout strategies

### Mobile (`<600`)

- Single column
- Bottom nav
- Sheets for filters
- Edge-to-edge media on PDP / marketing
- Margins 16

### Tablet (`600–1239`)

- Optional 2-column category + rail
- Nav rail or bottom nav (app-specific)
- Margins 24
- Split view: list | detail for orders (admin/customer large)

### Desktop (`≥1240`)

- Admin: sidebar + content
- Customer web (if any): max content width 1200–1280 centered; avoid ultra-stretched cards
- Margins 32
- Hover affordances enabled

### Large desktop (`≥1440`)

- Admin: secondary tools panel optional
- Wider data tables; pin columns

### Ultra wide (`≥1920`)

- Do not linearly stretch UI chrome
- Cap canvas; use whitespace or extra columns for dashboards
- Max readable measure ~72ch for prose

---

## Adaptive shells

| Width | Customer | Courier | Warehouse | Admin |
|-------|----------|---------|-----------|-------|
| xs–sm | BottomNav | BottomNav / map | Task chrome | Hamburger → drawer |
| md | Rail or BottomNav | Rail | Split queue/detail | Collapsed sidebar |
| lg+ | (web) top+ | — | — | Full sidebar |

`NxAdaptiveScaffold` switches structure from breakpoints + app density.

---

## Grid

- Customer product grid: 2 cols mobile; 3–4 tablet; 4–5 desktop
- Gutter: 12 mobile / 16 desktop
- Admin dashboards: 12-column grid; KPI row full width

## Pointer modalities

- Coarse: larger targets, no hover-only
- Fine: hover elevation, denser tables
