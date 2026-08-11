# 06 — Shadow & Elevation System

## Philosophy

Light mode: soft graphite-tinted shadows (not pure black).  
Dark mode: prefer **surface raised steps** over shadows; shadows only for true floaters (menus, dialogs).

---

## Elevation levels

| Level | Light shadow | Typical use |
|-------|--------------|-------------|
| `elevation.0` | none | Flat on canvas |
| `elevation.1` | `0 1 2 / 6%` + `0 1 1 / 4%` ink | Resting cards, inputs optional |
| `elevation.2` | `0 4 12 / 8%` | Raised cards, dropdowns, cart bar |
| `elevation.3` | `0 8 24 / 12%` | Bottom sheets, dialogs, FABs |
| `elevation.4` | `0 16 40 / 16%` | Command palette, critical overlays |

Ink color base: `#0B1214` at listed alphas (see `NxElevation`).

---

## Component mapping

| Component | Level (light) | Dark treatment |
|-----------|---------------|----------------|
| Product card resting | 0–1 + border.subtle | surfaceRaised + border |
| Product card pressed | 0 | same |
| Floating cart / CTA bar | 2 | raised + top border |
| FAB | 3 | 3 |
| Bottom sheet | 3 | 3 |
| Dialog | 3 | 3 |
| Navigation bottom bar | 1 or top hairline | hairline |
| Desktop sidebar | 0 + right border | border |
| Desktop floating panel | 2 | raised |
| Dropdown / popover | 2 | 2 |
| Toast / snackbar | 2 | 2 |

---

## Hover shadows (desktop / web)

- Resting → hover: +1 elevation step OR border.strong — not both dramatically
- Duration: `motion.fast` (120ms), curve standard
- Touch devices: no hover-dependent affordance as sole signal

---

## Rules

1. Never stack heavy shadows on nested cards
2. Prefer border for admin density over elevation
3. Map markers use separate map SDK shadows — do not double with Flutter elevation
4. Reduced transparency environments: increase border, decrease blur reliance
