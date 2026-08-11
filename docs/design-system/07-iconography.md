# 07 — Iconography

## Style

- Optical size tuned for 24px grid
- 1.75–2px stroke at 24px for **outlined**
- **Filled** for selected nav, status emphasis, active toggles
- Corner language matches radius.sm (slightly rounded joins — not soft blobs)
- Monochrome; color from semantic icon roles
- No emoji as UI icons

---

## Size scale

| Token | px | Use |
|-------|-----|-----|
| `icon.xs` | 12 | Inline badges, table affordances |
| `icon.sm` | 16 | Chips, input leading, meta |
| `icon.md` | 20 | Dense lists, courier |
| `icon.lg` | 24 | **Default** nav, buttons |
| `icon.xl` | 32 | Empty state inline, key actions |
| `icon.xxl` | 48 | Illustration-adjacent, onboarding |

Touch target ≥ 48×48 even when glyph is 24.

---

## Variants

| Variant | When |
|---------|------|
| Outlined | Default rest state |
| Filled | Selected, active, critical status |
| Rounded | Prefer over sharp; matches brand |
| Duotone | **Not used** in product UI (keep single ink) |

---

## Animated icons

Allowed only with purpose:

| Icon | Motion | Duration |
|------|--------|----------|
| Live / tracking pulse | Soft opacity + scale 1→1.06 | `etaBreathe` 1600ms loop |
| Cart add | Brief scale down-up | `micro` 180ms |
| Sync / offline | Rotate 360 | linear 900ms while syncing |
| Success check | Draw-on stroke | `short` 260ms |

Respect `prefers-reduced-motion` → static final frame.

---

## Catalog domains

### Status
`success`, `warning`, `danger`, `info`, `offline`, `sync`, `lock`, `verified`

### Navigation
`home`, `search`, `cart`, `orders`, `account`, `menu`, `back`, `close`, `more`

### Commerce
`tag`, `coupon`, `gift`, `wallet`, `loyalty`, `heart`, `heartFilled`, `star`, `share`, `filter`, `sort`, `barcode`, `qr`

### Courier
`bike`, `bag`, `navigate`, `arrived`, `handoff`, `earnings`, `support`, `sos`

### Warehouse
`scan`, `bin`, `pallet`, `pack`, `stage`, `count`, `exception`, `printer`

### Admin
`dashboard`, `city`, `store`, `users`, `flags`, `audit`, `settings`, `command`

---

## Rules

1. Shipping icon set as SVG sprite / Flutter IconData from single package `nexora_icons` (future) or `nexora_design/icons`
2. Do not mix Material Icons with custom set on the same screen without optical size audit
3. Status color + icon both required — never color alone
4. Mirror icons for RTL where directionality matters (`back`, `navigate`)
