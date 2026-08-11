# 03 — Typography

## Font stack

| Role | Family | Fallback |
|------|--------|----------|
| Display / headlines | **Satoshi** | ui-rounded, system-ui, Segoe UI, Roboto, sans-serif |
| Body / UI | **Geist** | system-ui, Segoe UI, Roboto, sans-serif |
| Mono / codes | **Geist Mono** | ui-monospace, Cascadia Mono, Consolas, monospace |

Do **not** use Inter, Roboto, or Arial as primary brand faces.

---

## Scale

| Token | Size / LH | Weight | Tracking | Family | Use |
|-------|-----------|--------|----------|--------|-----|
| `display.xl` | 40 / 48 | 700 | −2% | Satoshi | Marketing hero only |
| `display.lg` | 32 / 40 | 700 | −2% | Satoshi | Landing section titles |
| `headline.lg` | 28 / 36 | 700 | −1.5% | Satoshi | Screen titles (rare mobile) |
| `headline.md` | 24 / 32 | 600 | −1% | Satoshi | Page titles |
| `headline.sm` | 20 / 28 | 600 | −1% | Satoshi | Section titles |
| `title.lg` | 18 / 26 | 600 | 0 | Geist | Card titles large |
| `title.md` | 16 / 24 | 600 | 0 | Geist | List titles |
| `title.sm` | 14 / 20 | 600 | 0 | Geist | Compact titles |
| `body.lg` | 16 / 24 | 400 | 0 | Geist | Comfortable reading |
| `body.md` | 14 / 20 | 400 | 0 | Geist | **Default body** |
| `body.sm` | 13 / 18 | 400 | 0 | Geist | Dense / warehouse |
| `caption.md` | 12 / 16 | 400 | +1% | Geist | Meta |
| `caption.sm` | 11 / 14 | 500 | +2% | Geist | Badges micro |
| `overline` | 11 / 14 | 600 | +8% | Geist | Section labels (UPPER optional via style) |
| `button.lg` | 16 / 24 | 600 | 0 | Geist | Large CTA |
| `button.md` | 14 / 20 | 600 | 0 | Geist | Default button |
| `button.sm` | 13 / 18 | 600 | 0 | Geist | Compact button |
| `nav.md` | 12 / 16 | 500 | 0 | Geist | Bottom nav labels |
| `price.lg` | 20 / 28 | 700 | −1% | Geist **tabular** | PDP / checkout |
| `price.md` | 16 / 24 | 700 | 0 | Geist tabular | Cards |
| `price.sm` | 14 / 20 | 600 | 0 | Geist tabular | Compact |
| `eta.md` | 16 / 24 | 700 | 0 | Geist tabular | ETA values |
| `table.cell` | 13 / 18 | 400 | 0 | Geist tabular | Admin grids |
| `dash.kpi` | 32 / 40 | 700 | −1% | Geist tabular | Dashboard KPIs |

---

## Role mapping

| UI role | Token |
|---------|-------|
| Display | display.* |
| Headline | headline.* |
| Title | title.* |
| Subtitle | title.sm + text.secondary OR body.md secondary |
| Body | body.* |
| Caption | caption.* |
| Overline | overline |
| Button text | button.* |
| Navigation text | nav.md |
| Dashboard typography | dash.kpi + headline.sm + table.cell |
| Table typography | table.cell; headers = title.sm |

---

## Responsive typography

| Breakpoint | Adjustment |
|------------|------------|
| Mobile `<600` | Cap display at `display.lg`; prefer `headline.md` for pages |
| Tablet `600–1239` | Allow `headline.lg` |
| Desktop `≥1240` | `display.*` OK on marketing; admin stays dense (`table.cell`) |
| Ultra-wide `≥1920` | Do not scale type with viewport width; scale **layout columns** instead |

Fluid type is **not** used for product UI — discrete steps only (predictable ops glanceability).

---

## Font scaling (Dynamic Type / user a11y)

- Support system text scale 100%–200%
- Layouts must not clip at 135%
- At ≥150%: collapse secondary chrome; keep primary CTA + prices readable
- Tabular figures remain tabular under scale
- Hard-clipping truncated prices forbidden — wrap or scroll region

---

## Rules

1. Prices, ETAs, quantities, SKUs, order codes → **tabular figures**
2. Never bold entire paragraphs — use hierarchy
3. One display/headline per viewport section
4. Admin: prefer Geist throughout; Satoshi only for rare marketing-adjacent empty states
5. All user-facing strings via l10n — type styles applied in widgets
