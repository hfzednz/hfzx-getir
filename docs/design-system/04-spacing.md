# 04 — Spacing System

## Grid

**Base unit: 4pt.** All spacing tokens are multiples of 4 (plus half-step `2` for hairline optical adjustments only).

| Token | px |
|-------|-----|
| `space.0` | 0 |
| `space.0_5` | 2 |
| `space.1` | 4 |
| `space.2` | 8 |
| `space.3` | 12 |
| `space.4` | 16 |
| `space.5` | 20 |
| `space.6` | 24 |
| `space.8` | 32 |
| `space.10` | 40 |
| `space.12` | 48 |
| `space.16` | 64 |
| `space.20` | 80 |
| `space.24` | 96 |

---

## Usage recipes

| Concern | Mobile | Tablet | Desktop |
|---------|--------|--------|---------|
| Screen horizontal margin | 16 (`s4`) | 24 (`s6`) | 32 (`s8`) |
| Section spacing | 32 (`s8`) | 40 (`s10`) | 48 (`s12`) |
| Card internal padding | 12–16 | 16 | 16–20 |
| Card-to-card gap | 12 | 16 | 16 |
| List item vertical padding | 12 | 12 | 8–12 (density) |
| Form field stack gap | 16 | 16 | 16 |
| Button group gap | 8–12 | 12 | 12 |
| Bottom nav content inset | 8 + safe area | — | N/A (rail/sidebar) |
| Admin content padding | — | 24 | 24–32 |
| Admin sidebar width | — | 240 | 264 |

---

## Surface-specific

### Customer (comfortable)

- Home rails: horizontal peek padding 16; inter-card 12
- PDP: block sections 24 apart
- Sticky cart bar: 16 padding; 12 above home indicator

### Courier (compact)

- Task cards: 12 padding; 8 gaps
- Map chrome overlays: 12 from edges

### Warehouse (dense)

- Scan list rows: 8 vertical padding
- Action dock: 12 padding; large hit targets still ≥44

### Admin (dense)

- Table cell padding: 8×12
- Panel gaps: 16
- Do not use `s16+` between related controls — wastes density

---

## Rules

1. No magic numbers in widgets — use `NxSpacing.*`
2. Prefer padding on containers over margin on children for lists
3. Safe-area insets are additive to tokens, never replacements
4. Optical exceptions (icon alignment ±2) documented in component specs only
